package main

import (
	"flag"
	"fmt"
	"log/slog"
	"math"
	"os"
	"slices"

	"github.com/bit8bytes/gearberg/internal/orgs/settings"
)

type options struct {
	Version             bool
	LogLevel            logLevel
	Port                int
	BaseURL             string
	TLSMode             string
	DbDsn               string // SECRET
	StorageDSN          string
	SMTP                SMTP
	MaxOrgs             int
	MaxOrgCategories    int
	MaxOrgManufacturers int
	MaxOrgLocations     int
	MaxStorageBytes     int64
	DefaultCurrency     string
	DefaultVatRate      float64 // percentage e.g. 19.0
	DefaultTimezone     string
}

func registerCommonFlags(fs *flag.FlagSet, cfg *options) {
	fs.Var(&cfg.LogLevel, "log-level", "log level (debug|info|warn|error)")
	fs.StringVar(&cfg.DbDsn, "db-dsn", envOr("DB_DSN", "file:gearberg.db"), "database DSN")
}

func parseServeOptions(args []string) (*options, error) {
	cfg := &options{LogLevel: logLevel{level: slog.LevelInfo}}
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	registerCommonFlags(fs, cfg)
	fs.BoolVar(&cfg.Version, "version", false, "print version and exit")
	fs.IntVar(&cfg.Port, "port", 8080, "port to listen on")
	fs.StringVar(&cfg.BaseURL, "base-url", envOr("BASE_URL", ""), "base URL for link generation (e.g. https://example.com)")
	fs.StringVar(&cfg.TLSMode, "tls-mode", "off", "TLS mode (off|local)")
	fs.IntVar(&cfg.MaxOrgs, "max-orgs", 1, "maximum number of orgs allowed")
	fs.IntVar(&cfg.MaxOrgCategories, "max-categories", 25, "maximum number of equipment categories per org")
	fs.IntVar(&cfg.MaxOrgManufacturers, "max-manufacturers", 100, "maximum number of manufacturers per org")
	fs.IntVar(&cfg.MaxOrgLocations, "max-locations", 100, "maximum number of locations per org")
	fs.StringVar(&cfg.StorageDSN, "storage-dsn", envOr("STORAGE_DSN", "./var/data"), "storage backend DSN")
	fs.Int64Var(&cfg.MaxStorageBytes, "max-storage-bytes", 1<<30, "maximum storage bytes per org (default 1 GiB)")
	fs.StringVar(&cfg.DefaultCurrency, "default-currency", "EUR", "default currency for new org settings (ISO-4217)")
	fs.Float64Var(&cfg.DefaultVatRate, "default-vat-rate", 19.0, "default VAT rate for new org settings (percentage)")
	fs.StringVar(&cfg.DefaultTimezone, "default-timezone", "Europe/Berlin", "default timezone for new org settings (IANA)")

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("parseServeOptions: %w", err)
	}

	modes := []string{"off", "local"}
	if !slices.Contains(modes, cfg.TLSMode) {
		return nil, fmt.Errorf("tls-mode must be one of: %v", modes)
	}
	if cfg.TLSMode == "local" {
		return nil, fmt.Errorf("tls-mode 'local' is not supported yet")
	}
	if cfg.TLSMode == "off" {
		if err := validatePort(cfg.Port); err != nil {
			return nil, fmt.Errorf("invalid port: %w", err)
		}
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
	if !slices.Contains(settings.PermittedCurrencies, cfg.DefaultCurrency) {
		return fmt.Errorf("default-currency %q is not a permitted ISO-4217 code", cfg.DefaultCurrency)
	}
	if cfg.DefaultVatRate < 0 || cfg.DefaultVatRate > 100 {
		return fmt.Errorf("default-vat-rate must be between 0 and 100")
	}
	if !slices.Contains(settings.PermittedTimezones, cfg.DefaultTimezone) {
		return fmt.Errorf("default-timezone %q is not a permitted IANA timezone", cfg.DefaultTimezone)
	}
	return nil
}

// DefaultVatRateBasisPoints converts DefaultVatRate from a percentage to basis points (e.g. 19.0 → 1900).
func (cfg *options) DefaultVatRateBasisPoints() int64 {
	return int64(math.Round(cfg.DefaultVatRate * 100))
}

func parseVerifyOptions(args []string) (*options, error) {
	cfg := &options{LogLevel: logLevel{level: slog.LevelError}}
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	registerCommonFlags(fs, cfg)

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("parseVerifyOptions: %w", err)
	}

	return cfg, nil
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
