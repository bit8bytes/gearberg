package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"log/slog"
	"net/url"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/bit8bytes/gearberg/database"
	"github.com/bit8bytes/gearberg/database/migrations"
	"github.com/bit8bytes/gearberg/internal/inventory"
	"github.com/bit8bytes/gearberg/internal/orgs"
	"github.com/bit8bytes/gearberg/internal/orgs/categories"
	"github.com/bit8bytes/gearberg/internal/orgs/settings"
	"github.com/bit8bytes/gearberg/internal/storage"
	"github.com/bit8bytes/gearberg/templates"
	"github.com/bit8bytes/gearberg/templates/pages"
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
	}
}

func parseTemplates() (map[string]*template.Template, error) {
	base, err := template.New("root").Funcs(templateFuncs()).ParseFS(templates.EmbedFS, "layouts/root.tmpl", "components/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("base template: %w", err)
	}

	allPages := pages.All
	tmpls := make(map[string]*template.Template, len(allPages))
	for _, page := range allPages {
		t, err := pageTemplate(templates.EmbedFS, base, page)
		if err != nil {
			return nil, fmt.Errorf("page template %s: %w", page.File, err)
		}
		tmpls[page.File] = t
	}
	return tmpls, nil
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

type services struct {
	orgs                *orgs.Service
	orgsettings         *settings.Service
	equipmentcategories *categories.Service
	inventory           *inventory.Service
	storageManager      *storage.Manager
}

func setupServices(db *sql.DB, opts *options, logger *slog.Logger) (*services, error) {
	orgRepo := orgs.NewRepository(db)
	orgSvc := orgs.NewService(orgRepo, orgs.Options{MaxOrgs: opts.MaxOrgs})

	orgsettingsRepo := settings.NewRepository(db)
	orgsettingsSvc := settings.NewService(orgsettingsRepo)

	equipmentcategoriesRepo := categories.NewRepository(db)
	equipmentcategoriesSvc := categories.NewService(equipmentcategoriesRepo, categories.Options{MaxCategories: opts.MaxOrgCategories})

	inventoryRepo := inventory.NewRepository(db)
	inventorySvc := inventory.NewService(inventoryRepo, equipmentcategoriesSvc)

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
		orgs:                orgSvc,
		orgsettings:         orgsettingsSvc,
		equipmentcategories: equipmentcategoriesSvc,
		inventory:           inventorySvc,
		storageManager:      storageMgr,
	}, nil
}
