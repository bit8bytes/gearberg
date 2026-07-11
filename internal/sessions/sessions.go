// Package sessions provides utilities for storing and retrieving authentication
// and authorization metadata within a request context.
//
// This package is designed to support a denormalized security model where
// AccountID and OrganizationID are passed together to satisfy database queries
// without requiring complex joins.
package sessions

import (
	"context"
	"net/http"
)

// contextKey is a private type to prevent collisions with other context keys.
type contextKey string

// key is the unique identifier for session data within a context.
const key contextKey = "session"

// Session represents the authenticated state of a user for the duration
// of a single request. It contains the IDs necessary to perform
// direct-lookup authorization in the database.
type Session struct {
	// AccountID is the unique identifier of the authenticated user.
	AccountID string
}

// NewContext returns a new context derived from ctx that contains the Session s.
// Use this in middleware to inject session data after successful authentication.
func NewContext(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, key, s)
}

// MustFromRequest extracts the Session from an http.Request context.
//
// If no session is found, it panics. Use this in handlers or logic layers
// where authentication is guaranteed by middleware.
func MustFromRequest(r *http.Request) Session {
	s, ok := r.Context().Value(key).(Session)
	if !ok {
		panic("sessions: no session in context")
	}
	return s
}
