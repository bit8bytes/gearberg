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
	"flag"
	"fmt"
	"log/slog"
	"slices"
)

type options struct {
	Version  bool
	LogLevel logLevel
	Port     int
	TLSMode  string
}

func parseOptions() (*options, error) {
	cfg := &options{LogLevel: logLevel{level: slog.LevelError}}
	flag.Var(&cfg.LogLevel, "log-level", "log level (debug|info|warn|error)")
	flag.BoolVar(&cfg.Version, "version", false, "print version and exit")
	flag.IntVar(&cfg.Port, "port", 8080, "port to listen on")
	flag.StringVar(&cfg.TLSMode, "tls-mode", "off", "TLS mode (off|local)")
	flag.Parse()

	modes := []string{"off"}
	if !slices.Contains(modes, cfg.TLSMode) {
		return nil, fmt.Errorf("tls-mode must be one of: %v", modes)
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
