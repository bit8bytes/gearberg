package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/bit8bytes/gearberg/internal/accounts"
	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/internal/templates/fragments"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
	"github.com/bit8bytes/gearberg/pkg/htmx"
	"github.com/segmentio/ksuid"
)

func (app *application) getSignIn(w http.ResponseWriter, r *http.Request) *httperr.Error {
	tmplData := app.html.TemplateData(r)
	tmplData.Form = &accounts.SignInForm{}
	tmplData.Data = app.loginData()
	return app.html.Render(w, r, http.StatusOK, pages.SignIn, tmplData)
}

func (app *application) postSignIn(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()

	reRender := func(formWithErrors *accounts.SignInForm) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = formWithErrors
		data.Data = app.loginData()
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.SignIn, data)
	}

	form, err := accounts.ParseSignInForm(r)
	if err != nil {
		return reRender(&form)
	}

	if !form.Validate() {
		return reRender(&form)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	accountID, err := app.services.accounts.SignIn(ctx, accounts.SignInParams{
		Email:    form.Email,
		Password: form.Password,
	})
	if err != nil {
		form.AddError("email", "Invalid email or password.")
		return reRender(&form)
	}

	sessionSetAccountID(r.Context(), app.session, accountID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func (app *application) getSignUp(w http.ResponseWriter, r *http.Request) *httperr.Error {
	tmplData := app.html.TemplateData(r)
	tmplData.Form = &accounts.SignUpForm{}
	return app.html.Render(w, r, http.StatusOK, pages.SignUp, tmplData)
}

func (app *application) postSignUp(w http.ResponseWriter, r *http.Request) *httperr.Error {
	reRender := func(formWithErrors *accounts.SignUpForm) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = formWithErrors
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.SignUp, data)
	}

	form, err := accounts.ParseSignUpForm(r)
	if err != nil {
		return reRender(&form)
	}

	if !form.Validate() {
		return reRender(&form)
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
		return reRender(&form)
	}

	sessionSetAccountID(r.Context(), app.session, accountID)

	_, _ = app.services.orgsettings.Create(ctx, ksuid.New().String(), orgID)

	http.Redirect(w, r, "/orgs/"+orgID+"/equipment", http.StatusSeeOther)
	return nil
}

func (app *application) getForgotPassword(w http.ResponseWriter, r *http.Request) *httperr.Error {
	tmplData := app.html.TemplateData(r)
	tmplData.Form = &accounts.ForgotPasswordForm{}
	return app.html.Render(w, r, http.StatusOK, pages.ForgotPassword, tmplData)
}

func (app *application) postForgotPassword(w http.ResponseWriter, r *http.Request) *httperr.Error {
	reRender := func(formWithErrors *accounts.ForgotPasswordForm) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = formWithErrors
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.ForgotPassword, data)
	}

	form, err := accounts.ParseForgotPasswordForm(r)
	if err != nil {
		return reRender(&form)
	}

	if !form.Validate() {
		return reRender(&form)
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
	tmplData.Form = &accounts.ForgotPasswordForm{}
	return app.html.Render(w, r, http.StatusOK, pages.ForgotPasswordSuccess, tmplData)
}

func (app *application) getResetPassword(w http.ResponseWriter, r *http.Request) *httperr.Error {
	tmplData := app.html.TemplateData(r)
	tmplData.Form = &accounts.ResetPasswordForm{Token: r.URL.Query().Get("token")}
	return app.html.Render(w, r, http.StatusOK, pages.ResetPassword, tmplData)
}

func (app *application) postResetPassword(w http.ResponseWriter, r *http.Request) *httperr.Error {
	reRender := func(formWithErrors *accounts.ResetPasswordForm) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = formWithErrors
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.ResetPassword, data)
	}

	form, err := accounts.ParseResetPasswordForm(r)
	if err != nil {
		return reRender(&form)
	}

	if !form.Validate() {
		return reRender(&form)
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
		return reRender(&form)
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
