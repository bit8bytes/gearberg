package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"slices"
)

type options struct {
	Version  bool
	LogLevel logLevel
	Port     int
	TLSMode  string
	DbDsn    string // SECRET
}

// parseOptions parses command-line flags and returns Options.
func parseOptions() (*options, error) {
	cfg := &options{}

	flag.BoolVar(&cfg.Version, "version", false, "show application version")
	flag.Var(&cfg.LogLevel, "log-level", "Log level (debug|info|warn|error)")
	flag.IntVar(&cfg.Port, "port", 8080, "Port to listen on (without -domain)")

	flag.StringVar(&cfg.TLSMode, "tls-mode", "off", "TLS mode (off|local)")

	flag.StringVar(&cfg.DbDsn, "db-dsn", envOr("DB_DSN", "file:gearberg.db"), "database dsn")

	flag.Parse()

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
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
