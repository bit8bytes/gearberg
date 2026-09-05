// Copyright (C) 2026 Tobias Gleiter
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/alexedwards/scs/v2"
	"github.com/bit8bytes/gearberg/internal/accounts"
	"github.com/bit8bytes/gearberg/internal/categories"
	"github.com/bit8bytes/gearberg/internal/credentials"
	"github.com/bit8bytes/gearberg/internal/database"
	"github.com/bit8bytes/gearberg/internal/database/migrations"
	"github.com/bit8bytes/gearberg/internal/equipment"
	"github.com/bit8bytes/gearberg/internal/federated"
	imports "github.com/bit8bytes/gearberg/internal/import"
	"github.com/bit8bytes/gearberg/internal/locations"
	"github.com/bit8bytes/gearberg/internal/manufacturers"
	"github.com/bit8bytes/gearberg/internal/orgs"
	"github.com/bit8bytes/gearberg/internal/orgs/settings"
	"github.com/bit8bytes/gearberg/internal/storage"
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
		"inPath": func(currentPath, targetPath string) bool {
			return currentPath == targetPath || strings.HasPrefix(currentPath, targetPath+"/")
		},
	}
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
	mgr.Cookie.Name = "gearberg"
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
	PopString(ctx context.Context, key string) string
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

func (s scsSession) PopString(ctx context.Context, key string) string {
	return s.mgr.PopString(ctx, key)
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

func sessionAccountID(ctx context.Context, s sessionManager) string {
	return s.GetString(ctx, accounts.Key.String())
}

func sessionSetAccountID(ctx context.Context, s sessionManager, id string) {
	s.Put(ctx, accounts.Key.String(), id)
}

func sessionOIDCState(ctx context.Context, s sessionManager) string {
	return s.GetString(ctx, oidcStateKey)
}

func sessionSetOIDCState(ctx context.Context, s sessionManager, state string) {
	s.Put(ctx, oidcStateKey, state)
}

// oidcProvider holds the runtime config for a single OIDC provider.
type oidcProvider struct {
	oauth2Config         oauth2.Config
	verifier             *gooidc.IDTokenVerifier
	requireEmailVerified bool
}

// setupOIDCProvider initialises the OIDC provider from the given config.
// Returns nil when the provider is not configured so the caller can gate on it.
func setupOIDCProvider(ctx context.Context, cfg OIDCProvider, redirectURL string) (*oidcProvider, error) {
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
		verifier:             provider.Verifier(&gooidc.Config{ClientID: cfg.ClientID}),
		requireEmailVerified: cfg.RequireEmailVerified,
	}, nil
}

// setupOIDCProviders initialises all configured OIDC providers and returns a name→provider map.
func setupOIDCProviders(ctx context.Context, cfgs OIDCProviderMap, baseURL string) (map[string]*oidcProvider, error) {
	result := make(map[string]*oidcProvider, len(cfgs))
	for name, cfg := range cfgs {
		p, err := setupOIDCProvider(ctx, cfg, baseURL+"/auth/oidc/"+name+"/callback")
		if err != nil {
			return nil, fmt.Errorf("oidc-provider %s: %w", name, err)
		}
		result[name] = p
	}
	return result, nil
}

func setupMailer(cfg *options, _ *slog.Logger) mailer {
	if cfg.SMTP.Host == "" {
		return mailerpkg.NoopMailer{}
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
	equipmentImports    *imports.Service
	storageManager      *storage.Manager
}

// mailer defines the interface for sending emails.
type mailer interface {
	Mail(ctx context.Context, to, subject, body string) error
}

func setupServices(db *sql.DB, opts *options, logger *slog.Logger, m mailer) (*services, error) {
	orgRepo := orgs.NewRepository(db, int64(opts.Limits.MaxOrgs))
	orgSvc := orgs.NewService(db, orgRepo, orgRepo)

	accountRepo := accounts.NewRepository(db)
	credRepo := credentials.NewRepository(db)
	tokenRepo := tokens.NewRepository(db)
	federatedRepo := federated.NewRepository(db)

	accountSvc := accounts.NewService(db, &credentials.Password{}, accountRepo, credRepo, orgRepo, tokenRepo, m, federatedRepo)

	orgsettingsRepo := settings.NewRepository(db)
	orgsettingsSvc := settings.NewService(orgsettingsRepo)

	equipmentcategoriesRepo := categories.NewRepository(db)
	equipmentcategoriesSvc := categories.NewService(equipmentcategoriesRepo, categories.Options{MaxCategories: opts.Limits.MaxOrgCategories})

	manufacturersRepo := manufacturers.NewRepository(db)
	manufacturersSvc := manufacturers.NewService(manufacturersRepo, manufacturers.Options{MaxManufacturers: opts.Limits.MaxOrgManufacturers})

	locationsRepo := locations.NewRepository(db)
	locationsSvc := locations.NewService(locationsRepo, locations.Options{MaxLocations: opts.Limits.MaxOrgLocations})

	inventoryRepo := equipment.NewRepository(db)
	inventorySvc := equipment.NewService(inventoryRepo, db)

	equipmentImportPipeline := equipment.NewImportPipeline(inventorySvc, equipmentcategoriesSvc, manufacturersSvc, locationsSvc)

	importsRepo := imports.NewRepository(db)
	equipmentImportsSvc := imports.NewService(db, importsRepo, equipmentImportPipeline, equipmentImportPipeline, "equipment")

	store, err := storage.Open("local", opts.StorageDSN, logger)
	if err != nil {
		return nil, fmt.Errorf("setupServices: open storage: %w", err)
	}
	storageMgr := storage.NewManager(
		map[string]*storage.Store{"local": store},
		"local",
		opts.Limits.MaxStorageBytes,
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
