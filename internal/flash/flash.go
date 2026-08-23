// Package flash provides one-shot notification messages stored in the session
// and rendered as toasts on the next page load (post-redirect-get pattern).
package flash

import "context"

// Type controls the DaisyUI alert style of a notification.
type Type string

// Notification types mapped to DaisyUI alert styles.
const (
	Success Type = "success"
	Error   Type = "error"
)

// Flash is a one-shot notification message displayed as a toast.
type Flash struct {
	Message string
	Type    Type
}

// Manager stores and retrieves flash notifications across requests.
type Manager interface {
	// Put stores a flash notification to be shown on the next page load.
	Put(ctx context.Context, msg string, t Type)
	// Pop retrieves and removes the flash notification.
	// Returns nil if no flash is set.
	Pop(ctx context.Context) *Flash
}
