package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"slices"
)

type options struct {
	Version         bool
	LogLevel        logLevel
	Port            int
	TLSMode         string
	DbDsn           string // SECRET
	StorageDSN      string
	MaxCompanies    int
	MaxCategories   int
	MaxStorageBytes int64
}

func registerCommonFlags(fs *flag.FlagSet, cfg *options) {
	fs.Var(&cfg.LogLevel, "log-level", "log level (debug|info|warn|error)")
	fs.StringVar(&cfg.DbDsn, "db-dsn", envOr("DB_DSN", "file:gearberg.db"), "database DSN")
}

func parseServeOptions(args []string) (*options, error) {
	cfg := &options{}
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	registerCommonFlags(fs, cfg)
	fs.BoolVar(&cfg.Version, "version", false, "print version and exit")
	fs.IntVar(&cfg.Port, "port", 8080, "port to listen on")
	fs.StringVar(&cfg.TLSMode, "tls-mode", "off", "TLS mode (off|local)")
	fs.IntVar(&cfg.MaxCompanies, "max-companies", 1, "maximum number of companies allowed")
	fs.IntVar(&cfg.MaxCategories, "max-categories", 25, "maximum number of equipment categories per company")
	fs.StringVar(&cfg.StorageDSN, "storage-dsn", envOr("STORAGE_DSN", "./var/data"), "storage backend DSN")
	fs.Int64Var(&cfg.MaxStorageBytes, "max-storage-bytes", 1<<30, "maximum storage bytes per company (default 1 GiB)")

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

	if cfg.StorageDSN == "" {
		return nil, fmt.Errorf("storage-dsn is required")
	}
	if cfg.MaxCompanies <= 0 {
		return nil, fmt.Errorf("max companies must be greater than 0")
	}
	if cfg.MaxCategories <= 0 {
		return nil, fmt.Errorf("max categories must be greater than 0")
	}

	return cfg, nil
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
