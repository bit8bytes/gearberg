package main

import (
	"net/http"

	"github.com/bit8bytes/gearberg/internal/federated"
	pkgtokens "github.com/bit8bytes/gearberg/pkg/tokens"
)

// LoginData carries the set of enabled OIDC providers to the sign-in and
// sign-up templates so they can render the appropriate provider buttons.
type LoginData struct {
	AuthentikEnabled bool
}

func (app *application) loginData() LoginData {
	return LoginData{
		AuthentikEnabled: app.oidcAuthentik != nil,
	}
}

// getAuthAuthentik initiates the Authentik OIDC login flow. It generates a
// random state token, stores it in the session for CSRF verification, then
// redirects the browser to Authentik's authorization endpoint.
func (app *application) getAuthAuthentik(w http.ResponseWriter, r *http.Request) {
	if app.oidcAuthentik == nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	state := pkgtokens.Generate().Hex()
	sessionSetOIDCState(r.Context(), app.session, state)
	http.Redirect(w, r, app.oidcAuthentik.oauth2Config.AuthCodeURL(state), http.StatusSeeOther)
}

// getAuthAuthentikCallback handles the redirect from Authentik after the user
// authenticates. It verifies the CSRF state, exchanges the authorization code
// for tokens, validates the ID token, then signs in or JIT-provisions an account.
func (app *application) getAuthAuthentikCallback(w http.ResponseWriter, r *http.Request) {
	if app.oidcAuthentik == nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	ctx := r.Context()

	state := r.URL.Query().Get("state")
	if state == "" || state != sessionOIDCState(ctx, app.session) {
		app.logger.WarnContext(ctx, "oidc: state mismatch in authentik callback")
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	code := r.URL.Query().Get("code")
	token, err := app.oidcAuthentik.oauth2Config.Exchange(ctx, code)
	if err != nil {
		app.logger.ErrorContext(ctx, "oidc: code exchange failed", "error", err)
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		app.logger.ErrorContext(ctx, "oidc: id_token missing from token response")
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	idToken, err := app.oidcAuthentik.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		app.logger.ErrorContext(ctx, "oidc: id_token verification failed", "error", err)
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := idToken.Claims(&claims); err != nil {
		app.logger.ErrorContext(ctx, "oidc: claims extraction failed", "error", err)
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	if !claims.EmailVerified {
		app.logger.WarnContext(ctx, "oidc: rejected login with unverified email", "subject", idToken.Subject)
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	accountID, err := app.services.accounts.SignInOrCreateWithOIDC(ctx, federated.AuthentikProvider.ID(), idToken.Subject, claims.Email, claims.EmailVerified)
	if err != nil {
		app.logger.ErrorContext(ctx, "oidc: sign in or create failed", "error", err)
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	sessionSetAccountID(r.Context(), app.session, accountID)
	app.redirectAfterLogin(ctx, w, r, accountID)
}
