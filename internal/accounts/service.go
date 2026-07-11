package accounts

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/bit8bytes/gearberg/internal/credentials"
	"github.com/bit8bytes/gearberg/internal/database"
	"github.com/bit8bytes/gearberg/internal/orgs"
	"github.com/bit8bytes/gearberg/internal/tokens"
	pkgtokens "github.com/bit8bytes/gearberg/pkg/tokens"
	"github.com/segmentio/ksuid"
)

// Sentinel errors returned by account service operations.
var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired reset token")
	ErrNotFound           = errors.New("account not found")
	ErrSamePassword       = errors.New("new password must differ from the current password")
)

type passwordHasher interface {
	CreateHash(plaintext string) ([]byte, error)
	ComparePasswordAndHash(plaintext string, hash []byte) (bool, error)
}

type accountsRepository interface {
	Create(ctx context.Context, tx *sql.Tx, id, email string) (string, error)
	Delete(ctx context.Context, tx *sql.Tx, id string) error
	GetByAccountID(ctx context.Context, accountID string) (Record, error)
}

type credentialsRepository interface {
	Create(ctx context.Context, tx *sql.Tx, accountID string, typeID int64, secretData []byte) error
	GetByEmail(ctx context.Context, email string, typeID int64) (credentials.Record, error)
	GetByAccountID(ctx context.Context, accountID string, typeID int64) (credentials.Record, error)
	UpdatePassword(ctx context.Context, tx *sql.Tx, accountID string, hash string) error
}

type organizationsRepository interface {
	Create(ctx context.Context, tx *sql.Tx, id, displayName string) (string, error)
	CreateMember(ctx context.Context, tx *sql.Tx, accountID, organizationID string, roleID int64) error
	DeleteOwnedByAccount(ctx context.Context, tx *sql.Tx, accountID string) error
}

type tokensRepository interface {
	Create(ctx context.Context, tx *sql.Tx, p tokens.CreateParams) error
	Delete(ctx context.Context, tx *sql.Tx, accountID string, scopeID int64) error
	VerifyAndIncrement(ctx context.Context, tx *sql.Tx, hash []byte) (tokens.Record, error)
}

type mailer interface {
	Mail(ctx context.Context, to, subject, body string) error
}

// Service provides account creation and authentication operations.
type Service struct {
	db            *sql.DB
	password      passwordHasher
	accounts      accountsRepository
	credentials   credentialsRepository
	organizations organizationsRepository
	tokens        tokensRepository
	mailer        mailer
}

// NewService returns a new Service.
func NewService(
	db *sql.DB,
	hasher passwordHasher,
	accounts accountsRepository,
	creds credentialsRepository,
	organizations organizationsRepository,
	tokens tokensRepository,
	mailer mailer,
) *Service {
	return &Service{
		db:            db,
		password:      hasher,
		accounts:      accounts,
		credentials:   creds,
		organizations: organizations,
		tokens:        tokens,
		mailer:        mailer,
	}
}

// Get retrieves the account record for the given account ID.
func (s *Service) Get(ctx context.Context, accountID string) (Record, error) {
	record, err := s.accounts.GetByAccountID(ctx, accountID)
	if err != nil {
		return Record{}, fmt.Errorf("accounts.Get: %w", err)
	}
	return record, nil
}

// SignUpParams holds the fields required to create a new account.
type SignUpParams struct {
	Email          string
	Password       string
	RepeatPassword string
}

// SignUp creates a new account, credential, profile, and default organization atomically.
// It relies on the database UNIQUE constraint to reject duplicate emails rather
// than a separate existence check, eliminating the check-then-act race window.
func (s *Service) SignUp(ctx context.Context, params SignUpParams) (string, string, error) {
	fail := func(err error) (string, string, error) {
		return "", "", fmt.Errorf("accounts.SignUp: %w", err)
	}

	if params.Password != params.RepeatPassword {
		return fail(ErrInvalidPassword)
	}

	hash, err := s.password.CreateHash(params.Password)
	if err != nil {
		return fail(ErrInvalidPassword)
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fail(err)
	}
	defer func() { _ = tx.Rollback() }()

	accountID := ksuid.New().String()
	if _, err := s.accounts.Create(ctx, tx, accountID, params.Email); err != nil {
		return fail(err)
	}

	if err := s.credentials.Create(ctx, tx, accountID, credentials.PasswordType.ID(), hash); err != nil {
		return fail(err)
	}

	orgID := ksuid.New().String()
	if _, err := s.organizations.Create(ctx, tx, orgID, "My Organization"); err != nil {
		return fail(err)
	}

	if err := s.organizations.CreateMember(ctx, tx, accountID, orgID, orgs.OwnerRole.ID()); err != nil {
		return fail(err)
	}

	if err := tx.Commit(); err != nil {
		return fail(err)
	}

	return accountID, orgID, nil
}

// SignInParams holds the fields required to authenticate an account.
type SignInParams struct {
	Email    string
	Password string
}

// SignIn authenticates an account with email and password, returning the account ID.
func (s *Service) SignIn(ctx context.Context, params SignInParams) (string, error) {
	record, err := s.credentials.GetByEmail(ctx, params.Email, credentials.PasswordType.ID())
	if err != nil {
		return "", ErrInvalidCredentials
	}

	match, err := s.password.ComparePasswordAndHash(params.Password, []byte(record.SecretData))
	if err != nil || !match {
		return "", ErrInvalidCredentials
	}

	return record.AccountID, nil
}

// ForgotPassword looks up the account by email, generates a password-reset token,
// replaces any existing token for that account, and returns the plaintext token.
// Returns ErrNotFound if no user with the given email exists, ensuring tokens are
// only issued for registered accounts.
func (s *Service) ForgotPassword(ctx context.Context, email string) (string, error) {
	fail := func(err error) (string, error) {
		return "", fmt.Errorf("accounts.ForgotPassword: %w", err)
	}

	creds, err := s.credentials.GetByEmail(ctx, email, credentials.PasswordType.ID())
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return fail(ErrNotFound)
		}
		return fail(err)
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fail(err)
	}
	defer func() { _ = tx.Rollback() }()

	plaintext := pkgtokens.Generate().String()
	hash := pkgtokens.Hash(plaintext)

	if err := s.tokens.Delete(ctx, tx, creds.AccountID, tokens.PasswordReset.ID()); err != nil {
		return fail(err)
	}

	if err := s.tokens.Create(ctx, tx, tokens.CreateParams{
		AccountID: creds.AccountID,
		ScopeID:   tokens.PasswordReset.ID(),
		Hash:      hash,
		ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
	},
	); err != nil {
		return fail(err)
	}

	if err := tx.Commit(); err != nil {
		return fail(err)
	}

	return plaintext, nil
}

// SendPasswordReset sends a password reset email with a link built from resetURL.
func (s *Service) SendPasswordReset(ctx context.Context, email, resetURL string) error {
	subject := "Reset your password"
	body := "To reset your password, follow this link:\n\n" + resetURL +
		"\n\nThis link expires in 15 minutes." +
		"\n\nIf you did not request a password reset, you can safely ignore this email."
	if err := s.mailer.Mail(ctx, email, subject, body); err != nil {
		return fmt.Errorf("SendPasswordReset: %w", err)
	}
	return nil
}

// ResetPassword verifies the plaintext token, then updates the user's password.
// The token is deleted after a successful reset. It returns the account email
// so the caller can send a notification.
func (s *Service) ResetPassword(ctx context.Context, token, passwd string) (string, error) {
	fail := func(err error) (string, error) {
		return "", fmt.Errorf("accounts.ResetPassword: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fail(err)
	}
	defer func() { _ = tx.Rollback() }()

	tokenHash := pkgtokens.Hash(token)
	tok, err := s.tokens.VerifyAndIncrement(ctx, tx, tokenHash)
	if err != nil {
		return fail(ErrInvalidToken)
	}

	creds, err := s.credentials.GetByAccountID(ctx, tok.AccountID, credentials.PasswordType.ID())
	if err != nil {
		return fail(ErrInvalidToken)
	}

	// Reject if the new password is the same as the current one.
	if same, err := s.password.ComparePasswordAndHash(passwd, creds.SecretData); err == nil && same {
		return fail(ErrSamePassword)
	}

	newHash, err := s.password.CreateHash(passwd)
	if err != nil {
		return fail(err)
	}

	if err := s.credentials.UpdatePassword(ctx, tx, creds.AccountID, string(newHash)); err != nil {
		return fail(err)
	}

	if err := s.tokens.Delete(ctx, tx, tok.AccountID, tok.TokenScopeID); err != nil {
		return fail(err)
	}

	if err := tx.Commit(); err != nil {
		return fail(err)
	}

	accounts, err := s.accounts.GetByAccountID(ctx, creds.AccountID)
	if err != nil {
		return fail(err)
	}

	return accounts.Email, nil
}

// SendPasswordChangedNotification sends an alert email informing the user that
// their password was just changed. The caller should log but not surface any
// error returned, since the reset has already succeeded.
func (s *Service) SendPasswordChangedNotification(ctx context.Context, email string) error {
	subject := "Your password was changed"
	body := "Your gearberg password was just changed.\n\n" +
		"If you did not make this change, please contact support immediately."
	if err := s.mailer.Mail(ctx, email, subject, body); err != nil {
		return fmt.Errorf("SendPasswordChangedNotification: %w", err)
	}
	return nil
}

// Delete removes an account and all organizations where it is the sole owner.
func (s *Service) Delete(ctx context.Context, accountID string) error {
	fail := func(err error) error {
		return fmt.Errorf("accounts.Delete: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fail(err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.organizations.DeleteOwnedByAccount(ctx, tx, accountID); err != nil {
		return fail(err)
	}

	if err := s.accounts.Delete(ctx, tx, accountID); err != nil {
		return fail(err)
	}

	if err := tx.Commit(); err != nil {
		return fail(err)
	}

	return nil
}
