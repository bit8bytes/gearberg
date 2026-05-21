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
