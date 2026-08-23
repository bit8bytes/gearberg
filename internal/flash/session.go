package flash

import "context"

const (
	messageKey = "flash.message"
	typeKey    = "flash.type"
)

// sessionStore is the subset of the SCS session manager used by SessionManager.
// *scs.SessionManager satisfies this interface.
type sessionStore interface {
	Put(ctx context.Context, key string, val any)
	PopString(ctx context.Context, key string) string
}

// SessionManager is a flash.Manager backed by a session store.
type SessionManager struct {
	store sessionStore
}

// NewSessionManager returns a SessionManager backed by store.
func NewSessionManager(store sessionStore) *SessionManager {
	return &SessionManager{store: store}
}

// Put stores a flash notification to be shown on the next page load.
func (s *SessionManager) Put(ctx context.Context, msg string, t Type) {
	s.store.Put(ctx, messageKey, msg)
	s.store.Put(ctx, typeKey, string(t))
}

// Pop retrieves and removes the flash notification. Returns nil if no flash is set.
func (s *SessionManager) Pop(ctx context.Context) *Flash {
	msg := s.store.PopString(ctx, messageKey)
	if msg == "" {
		return nil
	}
	return &Flash{
		Message: msg,
		Type:    Type(s.store.PopString(ctx, typeKey)),
	}
}
