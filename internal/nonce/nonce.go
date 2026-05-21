// Package nonce provides per-request CSP nonce generation and context propagation.
package nonce

import "context"

type contextKey string

const key contextKey = "nonce_id"

// NewContext returns a new context with the given trace ID stored.
func NewContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, key, id)
}

// From returns the trace ID stored in ctx, or an empty string if none.
func From(ctx context.Context) Nonce {
	v, _ := ctx.Value(key).(string)
	return Nonce(v)
}

// Nonce is a per-request random value used as a CSP nonce.
type Nonce string

func (n *Nonce) String() string {
	return string(*n)
}
