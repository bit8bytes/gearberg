// Copyright 2026 filmlet. All rights reserved.

// Package migrations provides access to files embedded in the running Go program.
// It embeds all sql migration files for single binary deployments.
package migrations

import "embed"

// EmbedFS holds all embedded SQL migration files for single-binary deployments.
//go:embed "sqlite/*.sql"
var EmbedFS embed.FS
