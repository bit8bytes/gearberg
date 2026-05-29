package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"log/slog"
	"net/url"
	"path/filepath"
	"slices"
	"strings"

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
	base, err := template.New("root").Funcs(templateFuncs()).ParseFS(templates.EmbedFS, "layouts/landing.tmpl", "components/*.tmpl")
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
