package tokens

// Scope represents a token scope.
// The integer value matches the id stored in the token_scopes table,
// which is seeded at startup rather than via migrations.
type Scope int64

const (
	// PasswordReset is the scope for password reset tokens.
	PasswordReset Scope = 1
	// EmailVerification is the scope for email verification tokens.
	EmailVerification Scope = 2
)

// ID returns the database id for the token scope.
func (s Scope) ID() int64 { return int64(s) }

// String returns the name stored in the token_scopes table.
func (s Scope) String() string {
	switch s {
	case PasswordReset:
		return "password-reset"
	case EmailVerification:
		return "email-verification"
	default:
		return ""
	}
}
