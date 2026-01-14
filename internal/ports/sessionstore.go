package ports

import "time"

// SessionStore persists compact agent memory across runs.
// Implementations should be safe for concurrent reads and best-effort writes.
type SessionStore interface {
	Load(sessionName string) (*SessionState, error)
	Save(sessionName string, state *SessionState) error
}

// SessionState is intentionally compact: a rolling summary plus a small tail of recent lines.
// This keeps prompt sizes bounded while still allowing continuity.
type SessionState struct {
	Version   int               `json:"version"`
	Summary   string            `json:"summary"`
	Recent    []string          `json:"recent"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Tags      []string          `json:"tags,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}
