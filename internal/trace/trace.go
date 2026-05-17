// Package trace provides helpers for propagating trace IDs through context.
package trace

import "context"

type contextKey string

const key contextKey = "trace_id"

// NewContext returns a new context with the given trace ID stored.
func NewContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, key, id)
}

// FromContext returns the trace ID stored in ctx, or an empty string if none.
func FromContext(ctx context.Context) string {
	v, _ := ctx.Value(key).(string)
	return v
}
