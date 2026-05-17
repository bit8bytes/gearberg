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
	"strings"
	"time"

	"github.com/bit8bytes/gearberg/assets"
	"github.com/bit8bytes/gearberg/database"
	"github.com/bit8bytes/gearberg/database/migrations"
	"github.com/bit8bytes/gearberg/internal/companies"
	"github.com/bit8bytes/gearberg/internal/companies/categories"
	"github.com/bit8bytes/gearberg/internal/companies/settings"
	"github.com/bit8bytes/gearberg/internal/healthz"
	"github.com/bit8bytes/gearberg/internal/inventory"
	"github.com/bit8bytes/gearberg/templates"
	"github.com/bit8bytes/gearberg/templates/pages"
	"github.com/tobiasgleiter/forma"
	"github.com/tobiasgleiter/forma/adapters/formago"
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

func setupHandlers(logger *slog.Logger, db *sql.DB, cache map[string]*template.Template) http.Handler {
	mux := http.NewServeMux()

	healthz.NewHandler(revision, databaseVersion).Routes(mux)

	formaConfig := forma.DefaultConfig()
	formaConfig.Logger = logger
	formaConfig.ErrorTemplate = cache[pages.Error.File]
	m := forma.New(formago.New(mux), formaConfig)

	forma.Get(m, forma.Operation{
		Path:     "/",
		Template: cache[pages.NotFound.File],
	}, func(_ context.Context, _ *struct{}) (*struct{}, error) {
		return nil, nil
	})
	mux.Handle("GET /{$}", http.RedirectHandler("/companies", http.StatusSeeOther))
	mux.Handle("GET /dist/", assets.ServeStaticFiles())
	mux.Handle("GET /favicon.ico", http.RedirectHandler("/dist/images/favicon.ico", http.StatusMovedPermanently))

	companyRepo := companies.NewRepository(db)
	companySvc := companies.NewService(companyRepo)
	companyHandler := companies.NewHandler(companySvc, cache)
	companyHandler.Routes(m)

	csRepo := settings.NewRepository(db)
	csSvc := settings.NewService(csRepo)
	settings.NewHandler(csSvc, cache).Routes(m)

	ecRepo := categories.NewRepository(db)
	ecSvc := categories.NewService(ecRepo)
	ecHandler := categories.NewHandler(ecSvc, cache)
	ecHandler.Routes(m)

	inventoryService := inventory.NewService(ecSvc)
	inventoryHandler := inventory.NewHandler(inventoryService, cache)
	inventoryHandler.Routes(m)

	antiCSRF := http.NewCrossOriginProtection()
	logRequest := newRequestLogger(logger)
	recoverPanic := newPanicRecoverer(logger)

	return withTrace(
		recoverPanic.handler(
			logRequest.handler(
				withSecurityHeaders(
					withMaxBodySize(
						antiCSRF.Handler(mux))))))
}
