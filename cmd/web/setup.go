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
	"strings"
	"time"

	"github.com/bit8bytes/gearberg/database"
	"github.com/bit8bytes/gearberg/database/migrations"
	"github.com/bit8bytes/gearberg/internal/companies"
	"github.com/bit8bytes/gearberg/internal/companies/categories"
	"github.com/bit8bytes/gearberg/internal/companies/settings"
	"github.com/bit8bytes/gearberg/internal/inventory"
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
			for _, v := range vals[key] {
				if v == val {
					return true
				}
			}
			return false
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
	companies           *companies.Service
	companysettings     *settings.Service
	equipmentcategories *categories.Service
	inventory           *inventory.Service
}

func setupServices(db *sql.DB, opts *options) *services {
	companyRepo := companies.NewRepository(db)
	companySvc := companies.NewService(companyRepo, companies.Options{MaxCompanies: opts.MaxCompanies})

	companysettingsRepo := settings.NewRepository(db)
	companysettingsSvc := settings.NewService(companysettingsRepo)

	equipmentcategoriesRepo := categories.NewRepository(db)
	equipmentcategoriesSvc := categories.NewService(equipmentcategoriesRepo, categories.Options{MaxCategories: opts.MaxCategories})

	inventoryRepo := inventory.NewRepository(db)
	inventorySvc := inventory.NewService(inventoryRepo, equipmentcategoriesSvc)

	return &services{
		companies:           companySvc,
		companysettings:     companysettingsSvc,
		equipmentcategories: equipmentcategoriesSvc,
		inventory:           inventorySvc,
	}
}
