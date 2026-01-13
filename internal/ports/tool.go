package ports

import (
	"context"
	"encoding/json"
)

// Tool is a safe, read-only capability an agent can request.
// Tools must be deterministic and must not have side effects.
type Tool interface {
	Name() string
	Description() string
	InputSchema() string
	Run(ctx context.Context, input json.RawMessage) (string, error)
}
