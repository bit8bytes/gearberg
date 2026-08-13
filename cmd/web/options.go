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
	"crypto/tls"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type options struct {
	LogLevel            logLevel
	Port                int
	TLSMode             string
	TLSCertPath         string
	TLSKeyPath          string
	BaseURL             string
	DbDsn               string // SECRET
	StorageDSN          string
	SMTP                SMTP // has SECRET
	OIDCProviders       OIDCProviderMap
	MaxOrgs             int
	MaxOrgCategories    int
	MaxOrgManufacturers int
	MaxOrgLocations     int
	MaxStorageBytes     int64
}

func parseOptions(args []string) (*options, error) {
	cfg := &options{LogLevel: logLevel{level: slog.LevelError}, OIDCProviders: make(OIDCProviderMap)}
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Var(&cfg.LogLevel, "log-level", "log level (debug|info|warn|error)")
	fs.StringVar(&cfg.DbDsn, "db-dsn", envOr("DB_DSN", "file:/data/gearberg.db"), "database DSN")
	fs.IntVar(&cfg.Port, "port", 8080, "port to listen on")
	fs.StringVar(&cfg.BaseURL, "base-url", "localhost:8080", "base URL for link generation (e.g. https://example.com)")
	fs.StringVar(&cfg.TLSMode, "tls-mode", "off", "TLS mode (off|local)")
	fs.StringVar(&cfg.TLSCertPath, "tls-cert-path", "", "Path to TLS certificate file (required with -tls-mode=local)")
	fs.StringVar(&cfg.TLSKeyPath, "tls-key-path", "", "Path to TLS key file (required with -tls-mode=local)")
	fs.IntVar(&cfg.MaxOrgs, "max-orgs", 1, "maximum number of orgs allowed")
	fs.IntVar(&cfg.MaxOrgCategories, "max-categories", 25, "maximum number of equipment categories per org")
	fs.IntVar(&cfg.MaxOrgManufacturers, "max-manufacturers", 100, "maximum number of manufacturers per org")
	fs.IntVar(&cfg.MaxOrgLocations, "max-locations", 100, "maximum number of locations per org")
	fs.StringVar(&cfg.StorageDSN, "storage-dsn", envOr("STORAGE_DSN", "file:///data/uploads"), "storage backend DSN")
	fs.Int64Var(&cfg.MaxStorageBytes, "max-storage-bytes", 1<<30, "maximum storage bytes per org (default 1 GiB)")
	fs.Var(&cfg.OIDCProviders, "oidc-provider", "OIDC provider: name,issuer=URL,client-id=ID,client-secret=SECRET (repeatable)")
	fs.StringVar(&cfg.SMTP.Host, "smtp-host", "", "SMTP server hostname (omit to log emails only)")
	fs.IntVar(&cfg.SMTP.Port, "smtp-port", 587, "SMTP server port")
	fs.StringVar(&cfg.SMTP.Username, "smtp-username", "", "SMTP username")
	fs.StringVar(&cfg.SMTP.Password, "smtp-password", envOr("SMTP_PASSWORD", ""), "SMTP password (prefer SMTP_PASSWORD env var over flag)")
	fs.StringVar(&cfg.SMTP.From, "smtp-from", "", "from address for outgoing mail")

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("parseServeOptions: %w", err)
	}

	// Load OIDC providers from env vars matching OIDC_<NAME>_PROVIDER.
	// Flags take precedence — env vars only fill in providers not already set via flag.
	for _, env := range os.Environ() {
		k, v, _ := strings.Cut(env, "=")
		after, ok := strings.CutPrefix(k, "OIDC_")
		if !ok {
			continue
		}
		name, ok := strings.CutSuffix(after, "_PROVIDER")
		if !ok {
			continue
		}
		name = strings.ToLower(name)
		if _, exists := cfg.OIDCProviders[name]; exists {
			continue
		}
		if err := cfg.OIDCProviders.Set(v); err != nil {
			return nil, fmt.Errorf("env %s: %w", k, err)
		}
	}

	if err := cfg.validateTLS(); err != nil {
		return nil, err
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = fmt.Sprintf("http://localhost:%d", cfg.Port)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *options) validate() error {
	if cfg.BaseURL == "" {
		return fmt.Errorf("base-url is required")
	}
	if cfg.StorageDSN == "" {
		return fmt.Errorf("storage-dsn is required")
	}
	if cfg.MaxOrgs <= 0 {
		return fmt.Errorf("max orgs must be greater than 0")
	}
	if cfg.MaxOrgCategories <= 0 {
		return fmt.Errorf("max categories must be greater than 0")
	}
	if cfg.MaxOrgManufacturers <= 0 {
		return fmt.Errorf("max manufacturers must be greater than 0")
	}
	if cfg.MaxOrgLocations <= 0 {
		return fmt.Errorf("max locations must be greater than 0")
	}
	for name, p := range cfg.OIDCProviders {
		if err := p.Valid(); err != nil {
			return fmt.Errorf("oidc-provider %s: %w", name, err)
		}
	}
	return nil
}

func validatePort(port int) error {
	if port < 0 || port > 65535 {
		return fmt.Errorf("port is not in valid range of 0-65535")
	}
	return nil
}

// logLevel wraps slog.Level for flag parsing.
type logLevel struct {
	level slog.Level
}

// String returns the log level as a string.
func (l logLevel) String() string {
	switch l.level {
	case slog.LevelDebug:
		return "debug"
	case slog.LevelInfo:
		return "info"
	case slog.LevelWarn:
		return "warn"
	case slog.LevelError:
		return "error"
	default:
		return "info"
	}
}

// Set validates and sets the log level value.
func (l *logLevel) Set(value string) error {
	switch value {
	case "debug":
		l.level = slog.LevelDebug
	case "info":
		l.level = slog.LevelInfo
	case "warn":
		l.level = slog.LevelWarn
	case "error":
		l.level = slog.LevelError
	default:
		return fmt.Errorf("log-level must be one of: debug, info, warn, error")
	}
	return nil
}

// Level returns the slog.Level value.
func (l logLevel) Level() slog.Level {
	return l.level
}

// envOr returns the value of the environment variable key, or fallback if not set.
func envOr(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

type SMTP struct {
	Host     string
	Port     int
	Username string
	Password string // SECRET
	From     string
}

// Valid returns nil when SMTP is not configured (host empty), allowing log-only mode.
// When host is set, all required fields must be present.
func (s *SMTP) Valid() error {
	if s.Host == "" {
		return nil
	}
	if s.Port <= 0 || s.Port > 65535 {
		return fmt.Errorf("smtp port is not in valid range of 1-65535")
	}
	if s.Username == "" && s.Password != "" {
		return fmt.Errorf("smtp password requires a username")
	}
	if s.From == "" {
		return fmt.Errorf("smtp from email address cannot be empty")
	}
	return nil
}

type OIDCProvider struct {
	IssuerURL    string
	ClientID     string
	ClientSecret string // SECRET
}

// Configured reports whether this provider has been enabled via flags.
func (p *OIDCProvider) Configured() bool {
	return p.IssuerURL != ""
}

// OIDCProviderMap implements flag.Value to allow --oidc-provider to be repeated.
// Each value has the form: name,issuer=URL,client-id=ID,client-secret=SECRET.
type OIDCProviderMap map[string]OIDCProvider

func (m OIDCProviderMap) String() string { return "" }

func (m *OIDCProviderMap) Set(v string) error {
	idx := strings.IndexByte(v, ',')
	if idx <= 0 {
		return fmt.Errorf("oidc-provider: expected name,issuer=URL,client-id=ID,client-secret=SECRET")
	}
	name := v[:idx]
	p := OIDCProvider{}
	for kv := range strings.SplitSeq(v[idx+1:], ",") {
		k, val, ok := strings.Cut(kv, "=")
		if !ok {
			return fmt.Errorf("oidc-provider %s: malformed key=value %q", name, kv)
		}
		switch k {
		case "issuer":
			p.IssuerURL = val
		case "client-id":
			p.ClientID = val
		case "client-secret":
			p.ClientSecret = val
		default:
			return fmt.Errorf("oidc-provider %s: unknown key %q", name, k)
		}
	}
	if *m == nil {
		*m = make(OIDCProviderMap)
	}
	(*m)[name] = p
	return nil
}

// Valid returns nil when the provider is not configured (issuer empty).
// When issuer is set, client ID and secret must also be present.
func (p *OIDCProvider) Valid() error {
	if !p.Configured() {
		return nil
	}
	if p.ClientID == "" {
		return fmt.Errorf("client ID is required when issuer is set")
	}
	if p.ClientSecret == "" {
		return fmt.Errorf("client secret is required when issuer is set")
	}
	return nil
}

func (cfg *options) validateTLS() error {
	switch cfg.TLSMode {
	case "off":
		if err := validatePort(cfg.Port); err != nil {
			return fmt.Errorf("invalid port: %w", err)
		}
	case "local":
		if cfg.TLSCertPath == "" {
			return fmt.Errorf("-tls-cert-path is required with -tls-mode=local")
		}
		if cfg.TLSKeyPath == "" {
			return fmt.Errorf("-tls-key-path is required with -tls-mode=local")
		}
		if _, err := tls.LoadX509KeyPair(cfg.TLSCertPath, cfg.TLSKeyPath); err != nil {
			return fmt.Errorf("invalid TLS cert/key pair: %w", err)
		}
		if err := validatePort(cfg.Port); err != nil {
			return fmt.Errorf("invalid port: %w", err)
		}
	default:
		return fmt.Errorf("tls-mode must be one of: off, local")
	}
	return nil
}
