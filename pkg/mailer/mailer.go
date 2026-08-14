// Package mailer provides email sending implementations for production and development.
package mailer

import (
	"context"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
)

// SMTP sends emails via Go's net/smtp.
type SMTP struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// New creates a new SMTPMailer with the given SMTP configuration.
func New(host, username, password, from string, port int) *SMTP {
	return &SMTP{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
	}
}

// Mail sends a plain-text email to the given recipient.
func (m *SMTP) Mail(_ context.Context, to, subject, body string) error {
	addr := net.JoinHostPort(m.Host, strconv.Itoa(m.Port))

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s",
		m.From, to, subject, body)

	var auth smtp.Auth
	if m.Username != "" {
		auth = smtp.PlainAuth("", m.Username, m.Password, m.Host)
	}

	if err := smtp.SendMail(addr, auth, m.From, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("Mail: %w", err)
	}
	return nil
}

// NoopMailer is a no-op mailer that discards emails silently.
type NoopMailer struct{}

// Mail discards the email and returns nil.
func (NoopMailer) Mail(_ context.Context, _, _, _ string) error {
	return nil
}
