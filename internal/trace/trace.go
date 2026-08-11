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

// Package trace provides helpers for propagating trace IDs through context.
// A trace ID is a string token attached to a request at its entry point and
// carried through every layer via context.Context. Use [NewContext] to attach
// an ID and [From] to retrieve it for logging or outbound headers.
package trace

import "context"

type contextKey string

const key contextKey = "trace_id"

// NewContext returns a copy of ctx with id stored as the trace ID.
// Callers should generate id before the first handler runs, typically as a
// UUID or the value of an incoming X-Trace-ID / X-Request-ID header.
func NewContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, key, id)
}

// From returns the trace ID stored in ctx.
// It returns an empty Trace if no ID has been set.
func From(ctx context.Context) Trace {
	v, _ := ctx.Value(key).(string)
	return Trace(v)
}

// Trace is the string representation of a trace ID.
type Trace string

// String returns the underlying string value of t.
func (t *Trace) String() string {
	return string(*t)
}
