package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"slices"
	"strings"
	"time"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/alexedwards/scs/v2"
	"github.com/bit8bytes/gearberg/internal/accounts"
	"github.com/bit8bytes/gearberg/internal/credentials"
	"github.com/bit8bytes/gearberg/internal/database"
	"github.com/bit8bytes/gearberg/internal/database/migrations"
	"github.com/bit8bytes/gearberg/internal/equipment"
	"github.com/bit8bytes/gearberg/internal/equipment/categories"
	invimports "github.com/bit8bytes/gearberg/internal/equipment/imports"
	"github.com/bit8bytes/gearberg/internal/equipment/locations"
	"github.com/bit8bytes/gearberg/internal/equipment/manufacturers"
	"github.com/bit8bytes/gearberg/internal/federated"
	"github.com/bit8bytes/gearberg/internal/orgs"
	"github.com/bit8bytes/gearberg/internal/orgs/settings"
	"github.com/bit8bytes/gearberg/internal/storage"
	"github.com/bit8bytes/gearberg/internal/templates"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
	"github.com/bit8bytes/gearberg/internal/tokens"
	mailerpkg "github.com/bit8bytes/gearberg/pkg/mailer"
)

func setupLogger(l slog.Level) *slog.Logger {
	level := new(slog.LevelVar)
	level.Set(l)

	opts := &slog.HandlerOptions{
		AddSource:   false,
		Level:       level,
		ReplaceAttr: includeSourceFile,
	}

	return slog.New(slog.NewJSONHandler(log.Writer(), opts))
}

func includeSourceFile(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.SourceKey {
		source := a.Value.Any().(*slog.Source)
		source.File = filepath.Base(source.File)
	}
	return a
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"inQuery": func(vals url.Values, key, val string) bool {
			return slices.Contains(vals[key], val)
		},
		"inPath": func(currentPath, targetPath string) bool {
			return currentPath == targetPath || strings.HasPrefix(currentPath, targetPath+"/")
		},
		// unixDate formats a *int64 Unix timestamp as "2006-01-02", or returns "" for nil.
		"unixDate": func(v *int64) string {
			if v == nil {
				return ""
			}
			return time.Unix(*v, 0).UTC().Format("2006-01-02")
		},
	}
}

func parseTemplates() (*template.Template, map[string]*template.Template, error) {
	base, err := template.New("root").Funcs(templateFuncs()).ParseFS(templates.EmbedFS, "layouts/root.tmpl", "components/*.tmpl", "fragments/*.tmpl")
	if err != nil {
		return nil, nil, fmt.Errorf("base template: %w", err)
	}

	allPages := pages.All
	tmpls := make(map[string]*template.Template, len(allPages))
	for _, page := range allPages {
		t, err := pageTemplate(templates.EmbedFS, base, page)
		if err != nil {
			return nil, nil, fmt.Errorf("page template %s: %w", page.File, err)
		}
		tmpls[page.File] = t
	}
	return base, tmpls, nil
}

func pageTemplate(fsys fs.FS, base *template.Template, page pages.Page) (*template.Template, error) {
	t := template.Must(base.Clone())

	patterns := []string{page.Layout.File, page.File}
	if page.Layout.Partials != "" {
		patterns = append(patterns, page.Layout.Partials)
	}

	t, err := t.ParseFS(fsys, patterns...)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	if t.Lookup("page") == nil {
		return nil, fmt.Errorf("page block not defined: %s", page.File)
	}

	return t, nil
}

func setupDatabase(ctx context.Context, options *options) (*sql.DB, error) {
	dbCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	db, err := database.New(ctx, options.DbDsn)
	if err != nil {
		return nil, fmt.Errorf("open database failure: %w", err)
	}

	dbVersion, err := database.Migrate(dbCtx, db, migrations.EmbedFS)
	if err != nil {
		return nil, fmt.Errorf("migrate database failure: %w", err)
	}
	databaseVersion = fmt.Sprintf("%d", dbVersion)

	return db, nil
}

func setupSCS(db *sql.DB) (*scs.SessionManager, error) {
	mgr := scs.New()
	mgr.Lifetime = 30 * 24 * time.Hour
	mgr.Cookie.Name = "filmlet"
	store, err := database.SessionStore(db, time.Hour)
	if err != nil {
		return nil, fmt.Errorf("session store: %w", err)
	}
	mgr.Store = store
	mgr.HashTokenInStore = true
	return mgr, nil
}

// sessionManager abstracts the SCS session store for testing and dependency inversion.
type sessionManager interface {
	GetString(ctx context.Context, key string) string
	Put(ctx context.Context, key string, val any)
	Destroy(ctx context.Context) error
	LoadAndSave(next http.Handler) http.Handler
}

func setupSessionManager(mgr *scs.SessionManager) sessionManager {
	return scsSession{mgr: mgr}
}

// scsSession wraps *scs.SessionManager and implements sessionManager.
type scsSession struct {
	mgr *scs.SessionManager
}

func (s scsSession) GetString(ctx context.Context, key string) string {
	return s.mgr.GetString(ctx, key)
}

func (s scsSession) Put(ctx context.Context, key string, val any) {
	s.mgr.Put(ctx, key, val)
}

func (s scsSession) Destroy(ctx context.Context) error {
	if err := s.mgr.Destroy(ctx); err != nil {
		return fmt.Errorf("session destroy: %w", err)
	}
	return nil
}

func (s scsSession) LoadAndSave(next http.Handler) http.Handler {
	return s.mgr.LoadAndSave(next)
}

const oidcStateKey = "oidc_state"

func sessionAccountID(s sessionManager, ctx context.Context) string {
	return s.GetString(ctx, accounts.Key.String())
}

func sessionSetAccountID(s sessionManager, ctx context.Context, id string) {
	s.Put(ctx, accounts.Key.String(), id)
}

func sessionOIDCState(s sessionManager, ctx context.Context) string {
	return s.GetString(ctx, oidcStateKey)
}

func sessionSetOIDCState(s sessionManager, ctx context.Context, state string) {
	s.Put(ctx, oidcStateKey, state)
}

// oidcProvider holds the runtime config for a single OIDC provider.
type oidcProvider struct {
	oauth2Config oauth2.Config
	verifier     *gooidc.IDTokenVerifier
}

// setupOIDCProvider initialises the OIDC provider from the given config.
// Returns nil when the provider is not configured so the caller can gate on it.
func setupOIDCProvider(ctx context.Context, cfg OIDCProvider, redirectURL string) (*oidcProvider, error) {
	if !cfg.Configured() {
		return nil, nil
	}

	provider, err := gooidc.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("setupOIDCProvider: discover %s: %w", cfg.IssuerURL, err)
	}

	return &oidcProvider{
		oauth2Config: oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  redirectURL,
			Scopes:       []string{gooidc.ScopeOpenID, "email"},
		},
		verifier: provider.Verifier(&gooidc.Config{ClientID: cfg.ClientID}),
	}, nil
}

func setupMailer(cfg *options, log *slog.Logger) mailer {
	if cfg.SMTP.Host == "" {
		log.Warn("SMTP not configured, emails will be logged only")
		return mailerpkg.LogMailer{}
	}
	return mailerpkg.New(
		cfg.SMTP.Host,
		cfg.SMTP.Username,
		cfg.SMTP.Password,
		cfg.SMTP.From,
		cfg.SMTP.Port,
	)
}

type services struct {
	accounts            *accounts.Service
	orgs                *orgs.Service
	orgsettings         *settings.Service
	equipmentcategories *categories.Service
	manufacturers       *manufacturers.Service
	locations           *locations.Service
	equipment           *equipment.Service
	equipmentImports    *invimports.Service
	storageManager      *storage.Manager
}

// mailer defines the interface for sending emails.
type mailer interface {
	Mail(ctx context.Context, to, subject, body string) error
}

func setupServices(db *sql.DB, opts *options, logger *slog.Logger, m mailer) (*services, error) {
	orgRepo := orgs.NewRepository(db, int64(opts.MaxOrgs))
	orgSvc := orgs.NewService(db, orgRepo)

	accountRepo := accounts.NewRepository(db)
	credRepo := credentials.NewRepository(db)
	tokenRepo := tokens.NewRepository(db)
	federatedRepo := federated.NewRepository(db)

	accountSvc := accounts.NewService(db, &credentials.Password{}, accountRepo, credRepo, orgRepo, tokenRepo, m, federatedRepo)

	orgsettingsRepo := settings.NewRepository(db)
	orgsettingsSvc := settings.NewService(orgsettingsRepo)

	equipmentcategoriesRepo := categories.NewRepository(db)
	equipmentcategoriesSvc := categories.NewService(equipmentcategoriesRepo, categories.Options{MaxCategories: opts.MaxOrgCategories})

	manufacturersRepo := manufacturers.NewRepository(db)
	manufacturersSvc := manufacturers.NewService(manufacturersRepo, manufacturers.Options{MaxManufacturers: opts.MaxOrgManufacturers})

	locationsRepo := locations.NewRepository(db)
	locationsSvc := locations.NewService(locationsRepo, locations.Options{MaxLocations: opts.MaxOrgLocations})

	inventoryRepo := equipment.NewRepository(db)
	inventorySvc := equipment.NewService(inventoryRepo, db, equipmentcategoriesSvc, manufacturersSvc, locationsSvc)

	equipmentImportsRepo := invimports.NewRepository(db)
	equipmentImportsSvc := invimports.NewService(equipmentImportsRepo, db, inventorySvc, equipmentcategoriesSvc, manufacturersSvc, locationsSvc)

	store, err := storage.Open("local", opts.StorageDSN, logger)
	if err != nil {
		return nil, fmt.Errorf("setupServices: open storage: %w", err)
	}
	storageMgr := storage.NewManager(
		map[string]*storage.Store{"local": store},
		"local",
		opts.MaxStorageBytes,
		db,
	)

	return &services{
		accounts:            accountSvc,
		orgs:                orgSvc,
		orgsettings:         orgsettingsSvc,
		equipmentcategories: equipmentcategoriesSvc,
		manufacturers:       manufacturersSvc,
		locations:           locationsSvc,
		equipment:           inventorySvc,
		equipmentImports:    equipmentImportsSvc,
		storageManager:      storageMgr,
	}, nil
}
