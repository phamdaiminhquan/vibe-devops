package chat

// Role represents the author of a message in a conversation.
// This is kept small and stable to support future chat-mode.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message is a single conversational turn.
type Message struct {
	Role    Role
	Content string
}

// History is an ordered list of messages (oldest -> newest).
type History []Message
