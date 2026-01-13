package command

// Suggestion represents a command proposed by the agent.
// Keeping this structured now lets us grow safety and metadata later
// without changing the CLI contract.
type Suggestion struct {
	Command string

	// Unsafe can be set by future safety checks.
	Unsafe bool

	// Raw is the unprocessed model output (optional; useful for debugging).
	Raw string
}
