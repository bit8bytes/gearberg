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
	"errors"
	"net/http"
	"time"

	"github.com/bit8bytes/gearberg/internal/accounts"
	"github.com/bit8bytes/gearberg/internal/federated"
	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/internal/templates/fragments"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
	"github.com/bit8bytes/gearberg/pkg/htmx"
	pkgtokens "github.com/bit8bytes/gearberg/pkg/tokens"
)

// redirectAfterLogin sends the user to their single org's equipment page if they
// belong to exactly one org, otherwise to the organizations settings page.
func (app *application) redirectAfterLogin(ctx context.Context, w http.ResponseWriter, r *http.Request, accountID string) {
	orgs, err := app.services.orgs.List(ctx, accountID)
	if err != nil {
		// It's better just warn in logs than breaking the login flow.
		app.logger.Warn("redirectAfterLogin: list orgs", "error", err)
	}

	if len(orgs) == 1 {
		http.Redirect(w, r, "/orgs/"+orgs[0].ID+"/equipment", http.StatusSeeOther) //nolint:gosec
	} else {
		http.Redirect(w, r, "/orgs/pick", http.StatusSeeOther)
	}
}

func (app *application) getSignIn(w http.ResponseWriter, r *http.Request) *httperr.Error {
	tmplData := app.html.TemplateData(r)
	tmplData.Form = accounts.NewSignInForm()
	tmplData.Data = app.loginData()
	return app.html.Render(w, r, http.StatusOK, pages.SignIn, tmplData)
}

func (app *application) postSignIn(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()

	fail := func(formWithErrors *accounts.SignInForm) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = formWithErrors
		data.Data = app.loginData()
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.SignIn, data)
	}

	form, err := accounts.ParseSignInForm(r)
	if err != nil {
		return fail(&form)
	}

	if !form.Validate() {
		return fail(&form)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	accountID, err := app.services.accounts.SignIn(ctx, accounts.SignInParams{
		Email:    form.Email,
		Password: form.Password,
	})
	if err != nil {
		form.AddError("email", "Invalid email or password.")
		return fail(&form)
	}

	sessionSetAccountID(r.Context(), app.session, accountID)
	app.redirectAfterLogin(ctx, w, r, accountID)
	return nil
}

func (app *application) getSignUp(w http.ResponseWriter, r *http.Request) *httperr.Error {
	tmplData := app.html.TemplateData(r)
	tmplData.Form = accounts.NewSignUpForm()
	return app.html.Render(w, r, http.StatusOK, pages.SignUp, tmplData)
}

func (app *application) postSignUp(w http.ResponseWriter, r *http.Request) *httperr.Error {
	fail := func(formWithErrors *accounts.SignUpForm) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = formWithErrors
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.SignUp, data)
	}

	form, err := accounts.ParseSignUpForm(r)
	if err != nil {
		return fail(&form)
	}

	if !form.Validate() {
		return fail(&form)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	accountID, orgID, err := app.services.accounts.SignUp(ctx, accounts.SignUpParams{
		Email:          form.Email,
		Password:       form.Password,
		RepeatPassword: form.RepeatPassword,
	})
	if err != nil {
		switch {
		case errors.Is(err, accounts.ErrUserAlreadyExists) || errors.Is(err, accounts.ErrInvalidPassword):
			form.AddError("email", "Invalid email or password.")
		default:
			app.logger.ErrorContext(ctx, "sign up failed", "error", err)
			form.AddError("email", "Invalid email or password.")
		}
		return fail(&form)
	}

	sessionSetAccountID(r.Context(), app.session, accountID)

	_, _ = app.services.orgsettings.Create(ctx, orgID)

	http.Redirect(w, r, "/orgs/"+orgID+"/equipment", http.StatusSeeOther)
	return nil
}

func (app *application) getForgotPassword(w http.ResponseWriter, r *http.Request) *httperr.Error {
	tmplData := app.html.TemplateData(r)
	tmplData.Form = accounts.NewForgotPasswordForm()
	return app.html.Render(w, r, http.StatusOK, pages.ForgotPassword, tmplData)
}

func (app *application) postForgotPassword(w http.ResponseWriter, r *http.Request) *httperr.Error {
	fail := func(formWithErrors *accounts.ForgotPasswordForm) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = formWithErrors
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.ForgotPassword, data)
	}

	form, err := accounts.ParseForgotPasswordForm(r)
	if err != nil {
		return fail(&form)
	}

	if !form.Validate() {
		return fail(&form)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Always redirect to success to avoid leaking whether an email is registered.
	token, err := app.services.accounts.ForgotPassword(ctx, form.Email)
	if err != nil {
		if errors.Is(err, accounts.ErrNotFound) {
			app.logger.WarnContext(ctx, "password reset requested for unregistered email")
		} else {
			app.logger.ErrorContext(ctx, "failed to generate password reset token", "error", err)
		}
		http.Redirect(w, r, "/forgot-password/success", http.StatusSeeOther)
		return nil
	}

	resetURL := app.options.BaseURL + "/reset-password?token=" + token
	if err := app.services.accounts.SendPasswordReset(ctx, form.Email, resetURL); err != nil {
		app.logger.ErrorContext(ctx, "failed to send password reset email", "error", err)
	}

	http.Redirect(w, r, "/forgot-password/success", http.StatusSeeOther)
	return nil
}

func (app *application) getForgotPasswordSuccess(w http.ResponseWriter, r *http.Request) *httperr.Error {
	tmplData := app.html.TemplateData(r)
	tmplData.Form = accounts.NewForgotPasswordForm()
	return app.html.Render(w, r, http.StatusOK, pages.ForgotPasswordSuccess, tmplData)
}

func (app *application) getResetPassword(w http.ResponseWriter, r *http.Request) *httperr.Error {
	tmplData := app.html.TemplateData(r)
	tmplData.Form = accounts.NewResetPasswordForm(r.URL.Query().Get("token"))
	return app.html.Render(w, r, http.StatusOK, pages.ResetPassword, tmplData)
}

func (app *application) postResetPassword(w http.ResponseWriter, r *http.Request) *httperr.Error {
	fail := func(formWithErrors *accounts.ResetPasswordForm) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = formWithErrors
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.ResetPassword, data)
	}

	form, err := accounts.ParseResetPasswordForm(r)
	if err != nil {
		return fail(&form)
	}

	if !form.Validate() {
		return fail(&form)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	email, err := app.services.accounts.ResetPassword(ctx, form.Token, form.Password)
	if err != nil {
		switch {
		case errors.Is(err, accounts.ErrSamePassword):
			form.AddError("password", "New password must be different from your current password.")
		default:
			form.AddError("password", "Invalid or expired reset link.")
		}
		return fail(&form)
	}

	if err := app.services.accounts.SendPasswordChangedNotification(ctx, email); err != nil {
		app.logger.ErrorContext(ctx, "failed to send password changed notification", "error", err)
	}

	// Destroy any sessions if the user has one.
	// If no session exists, the session manager is a no-op.
	if err := app.session.Destroy(ctx); err != nil {
		app.logger.ErrorContext(ctx, "failed to destroy session after password reset", "error", err)
	}

	http.Redirect(w, r, "/signin", http.StatusSeeOther)
	return nil
}

func (app *application) postValidatePassword(w http.ResponseWriter, r *http.Request) *httperr.Error {
	if err := r.ParseForm(); err != nil {
		return nil
	}
	data := app.html.TemplateData(r)
	data.Form = accounts.ValidatePassword(r.PostForm.Get("password"))
	return app.html.RenderFragment(w, r, http.StatusOK, fragments.PasswordValidation, data)
}

func (app *application) postSignOut(w http.ResponseWriter, r *http.Request) *httperr.Error {
	if err := app.session.Destroy(r.Context()); err != nil {
		tmplData := app.html.TemplateData(r)
		return app.html.Render(w, r, http.StatusInternalServerError, pages.Error, tmplData)
	}

	url := "/signin"
	if htmx.IsRequest(r) {
		htmx.Redirect(w, r, url, http.StatusOK)
	} else {
		http.Redirect(w, r, url, http.StatusSeeOther)
	}
	return nil
}

// LoginData carries the set of enabled OIDC provider names to the sign-in and
// sign-up templates so they can render the appropriate provider buttons.
type LoginData struct {
	OIDCProviders []string
}

func (app *application) loginData() LoginData {
	names := make([]string, 0, len(app.oidcProviders))
	for name := range app.oidcProviders {
		names = append(names, name)
	}
	return LoginData{OIDCProviders: names}
}

// getAuthOIDC initiates the OIDC login flow for the named provider. It generates
// a random state token, stores it in the session for CSRF verification, then
// redirects the browser to the provider's authorization endpoint.
func (app *application) getAuthOIDC(w http.ResponseWriter, r *http.Request, name string) {
	p, ok := app.oidcProviders[name]
	if !ok {
		// Silent redirect: exposing an error leaks which provider names are valid.
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	state := pkgtokens.Generate().Hex()
	sessionSetOIDCState(r.Context(), app.session, state)
	http.Redirect(w, r, p.oauth2Config.AuthCodeURL(state), http.StatusSeeOther)
}

// getAuthOIDCCallback handles the redirect from an OIDC provider after the user
// authenticates. It verifies the CSRF state, exchanges the authorization code
// for tokens, validates the ID token, then signs in or JIT-provisions an account.
// Security-sensitive failures (unknown provider, state mismatch) redirect silently
// to avoid leaking information to potential attackers.
func (app *application) getAuthOIDCCallback(w http.ResponseWriter, r *http.Request, name string) *httperr.Error {
	p, ok := app.oidcProviders[name]
	if !ok {
		// Silent redirect: exposing an error leaks which provider names are valid.
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return nil
	}

	ctx := r.Context()

	state := r.URL.Query().Get("state")
	if state == "" || state != sessionOIDCState(ctx, app.session) {
		// Silent redirect: exposing an error leaks CSRF flow details.
		app.logger.WarnContext(ctx, "oidc: state mismatch", "provider", name)
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return nil
	}

	code := r.URL.Query().Get("code")
	token, err := p.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return httperr.InternalServerError(err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return httperr.InternalServerError(errors.New("oidc: id_token missing"))
	}

	idToken, err := p.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return httperr.InternalServerError(err)
	}

	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return httperr.InternalServerError(err)
	}

	if p.requireEmailVerified && !claims.EmailVerified {
		return &httperr.Error{
			Error:   errors.New("oidc: email not verified"),
			Message: "Your email address has not been verified by the identity provider. Please verify your email and try again.",
			Code:    http.StatusForbidden,
		}
	}

	providerID, known := federated.ByName(name)
	if !known {
		// Silent redirect: internal misconfiguration — no actionable information for the user.
		app.logger.ErrorContext(ctx, "oidc: unknown provider name", "provider", name)
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return nil
	}

	accountID, err := app.services.accounts.SignInOrCreateWithOIDC(ctx, providerID.ID(), idToken.Subject, claims.Email, claims.EmailVerified)
	if err != nil {
		return httperr.InternalServerError(err)
	}

	sessionSetAccountID(r.Context(), app.session, accountID)
	app.redirectAfterLogin(ctx, w, r, accountID)
	return nil
}
