package main

import (
	"context"
	"net/http"
	"time"

	"github.com/bit8bytes/gearberg/internal/accounts"
	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/internal/sessions"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
	"github.com/bit8bytes/gearberg/pkg/htmx"
)

type accountData struct {
	ID            string
	EmailVerified *time.Time
}

func (app *application) getAccount(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	accountID := sessionAccountID(app.session, ctx)
	record, err := app.services.accounts.Get(ctx, accountID)
	if err != nil {
		return &httperr.Error{
			Error:   err,
			Message: "Failed to load account.",
			Code:    http.StatusInternalServerError,
		}
	}

	tmplData := app.html.TemplateData(r)
	tmplData.Data = accountData{
		ID:            record.ID,
		EmailVerified: record.EmailVerified,
	}
	tmplData.Form = &accounts.Form{Email: record.Email}
	return app.html.Render(w, r, http.StatusOK, pages.SettingsAccount, tmplData)
}

func (app *application) deleteAccount(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()

	fail := func() *httperr.Error {
		tmplData := app.html.TemplateData(r)
		return app.html.Render(w, r, http.StatusInternalServerError, pages.Error, tmplData)
	}

	session := sessions.MustFromRequest(r)

	if err := app.services.accounts.Delete(ctx, session.AccountID); err != nil {
		return fail()
	}

	if err := app.session.Destroy(r.Context()); err != nil {
		return fail()
	}

	if htmx.IsRequest(r) {
		htmx.Redirect(w, r, "/", http.StatusOK)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	return nil
}
