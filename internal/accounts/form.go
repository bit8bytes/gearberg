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

// Package accounts defines form types and validation rules for account workflows.
package accounts

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bit8bytes/toolbox/validator"
	"github.com/ccojocar/zxcvbn-go"
)

const (
	passwordMinLength        = 12
	passwordMaxLength        = 128
	passwordMinStrengthScore = 3
)

// Form holds the parsed fields from a form submission.
type Form struct {
	Email string
	validator.Validator
}

// SignInForm holds the parsed fields from a sign-in submission.
type SignInForm struct {
	Email    string
	Password string
	validator.Validator
}

// ParseSignInForm reads and parses the sign-in form fields from the request.
func ParseSignInForm(r *http.Request) (SignInForm, error) {
	if err := r.ParseForm(); err != nil {
		return SignInForm{}, fmt.Errorf("accounts.ParseSignInForm: %w", err)
	}
	return SignInForm{
		Email:    strings.TrimSpace(r.PostForm.Get("email")),
		Password: r.PostForm.Get("password"),
	}, nil
}

// Validate checks that email and password are present and well-formed.
func (f *SignInForm) Validate() bool {
	f.Check(validator.NotBlank(f.Email), "email", "This field cannot be blank")
	f.Check(validator.Matches(f.Email, validator.EmailRX), "email", "This field must be a valid email address")
	f.Check(validator.NotBlank(f.Password), "password", "This field cannot be blank")
	return f.Valid()
}

// SignUpForm holds the parsed fields from a sign-up submission.
type SignUpForm struct {
	Email          string
	Password       string
	RepeatPassword string
	validator.Validator
}

// ParseSignUpForm reads and parses the sign-up form fields from the request.
func ParseSignUpForm(r *http.Request) (SignUpForm, error) {
	if err := r.ParseForm(); err != nil {
		return SignUpForm{}, fmt.Errorf("accounts.ParseSignUpForm: %w", err)
	}
	return SignUpForm{
		Email:          strings.TrimSpace(r.PostForm.Get("email")),
		Password:       r.PostForm.Get("password"),
		RepeatPassword: r.PostForm.Get("repeat_password"),
	}, nil
}

// Validate checks email format, password strength, length, and confirmation match.
func (f *SignUpForm) Validate() bool {
	f.Check(validator.NotBlank(f.Email), "email", "This field cannot be blank")
	f.Check(validator.Matches(f.Email, validator.EmailRX), "email", "This field must be a valid email address")

	f.Check(validator.NotBlank(f.Password), "password", "This field cannot be blank")
	f.Check(validator.MinChars(f.Password, passwordMinLength), "password",
		fmt.Sprintf("This field must be at least %v characters long", passwordMinLength))
	f.Check(validator.MaxChars(f.Password, passwordMaxLength),
		"password", fmt.Sprintf("This field must be less than %v characters long", passwordMaxLength))

	passwordStrength := zxcvbn.PasswordStrength(f.Password, nil)
	f.Check(passwordStrength.Score >= passwordMinStrengthScore, "password", "Password is too common")

	f.Check(validator.NotBlank(f.RepeatPassword), "repeat_password", "This field cannot be blank")
	f.Check(f.Password == f.RepeatPassword, "repeat_password", "Passwords do not match")
	return f.Valid()
}

// ForgotPasswordForm holds the parsed fields from a forgot-password submission.
type ForgotPasswordForm struct {
	Email string
	validator.Validator
}

// ParseForgotPasswordForm reads and parses the forgot-password form fields from the request.
func ParseForgotPasswordForm(r *http.Request) (ForgotPasswordForm, error) {
	if err := r.ParseForm(); err != nil {
		return ForgotPasswordForm{}, fmt.Errorf("accounts.ParseForgotPasswordForm: %w", err)
	}
	return ForgotPasswordForm{
		Email: strings.TrimSpace(r.PostForm.Get("email")),
	}, nil
}

// Validate checks that email is present and well-formed.
func (f *ForgotPasswordForm) Validate() bool {
	f.Check(validator.NotBlank(f.Email), "email", "This field cannot be blank")
	f.Check(validator.Matches(f.Email, validator.EmailRX), "email", "This field must be a valid email address")
	return f.Valid()
}

// ProfileForm holds form state for the profile settings page.
type ProfileForm struct {
	FirstName string
	LastName  string
	Phone     string
	validator.Validator
}

// ResetPasswordForm holds the parsed fields from a password-reset submission.
type ResetPasswordForm struct {
	Token          string
	Password       string
	RepeatPassword string
	validator.Validator
}

// ParseResetPasswordForm reads and parses the reset-password form fields from the request.
func ParseResetPasswordForm(r *http.Request) (ResetPasswordForm, error) {
	if err := r.ParseForm(); err != nil {
		return ResetPasswordForm{}, fmt.Errorf("accounts.ParseResetPasswordForm: %w", err)
	}
	return ResetPasswordForm{
		Token:          r.PostForm.Get("token"),
		Password:       r.PostForm.Get("password"),
		RepeatPassword: r.PostForm.Get("repeat_password"),
	}, nil
}

// PasswordValidationForm is used by the HTMX inline password validation endpoint.
// It has the same Password and Errors fields as SignUpForm and ResetPasswordForm,
// so the password-validation fragment template works with all three.
type PasswordValidationForm struct {
	Password string
	validator.Validator
}

// ValidatePassword validates a single password value and returns a form with any errors set.
func ValidatePassword(password string) *PasswordValidationForm {
	f := &PasswordValidationForm{Password: password}
	if password == "" {
		return f
	}
	f.Check(validator.MinChars(password, passwordMinLength), "password",
		fmt.Sprintf("Must be at least %d characters long", passwordMinLength))
	f.Check(validator.MaxChars(password, passwordMaxLength), "password",
		fmt.Sprintf("Must be less than %d characters long", passwordMaxLength))
	result := zxcvbn.PasswordStrength(password, nil)
	f.Check(result.Score >= passwordMinStrengthScore, "password", "Password is too weak or common")
	return f
}

// PasswordMinLength returns the minimum allowed password length for use in templates.
func (f *ResetPasswordForm) PasswordMinLength() int { return passwordMinLength }

// PasswordMaxLength returns the maximum allowed password length for use in templates.
func (f *ResetPasswordForm) PasswordMaxLength() int { return passwordMaxLength }

// Validate checks password strength, length, and confirmation match.
func (f *ResetPasswordForm) Validate() bool {
	f.Check(validator.NotBlank(f.Password), "password", "This field cannot be blank")
	f.Check(validator.MinChars(f.Password, passwordMinLength), "password",
		fmt.Sprintf("This field must be at least %v characters long", passwordMinLength))
	f.Check(validator.MaxChars(f.Password, passwordMaxLength),
		"password", fmt.Sprintf("This field must be less than %v characters long", passwordMaxLength))

	result := zxcvbn.PasswordStrength(f.Password, nil)
	f.Check(result.Score >= passwordMinStrengthScore, "password", "Password is too common")

	f.Check(validator.NotBlank(f.RepeatPassword), "repeat_password", "This field cannot be blank")
	f.Check(f.Password == f.RepeatPassword, "repeat_password", "Passwords do not match")
	return f.Valid()
}
