package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	ActionTypeTool   = "tool"
	ActionTypeDone   = "done"
	ActionTypeAnswer = "answer"
)

type Action struct {
	Type        string          `json:"type"`
	Tool        string          `json:"tool,omitempty"`
	Input       json.RawMessage `json:"input,omitempty"`
	Command     string          `json:"command,omitempty"`
	Explanation string          `json:"explanation,omitempty"`
}

func ParseAction(modelText string) (Action, error) {
	obj, err := extractFirstJSONObject(modelText)
	if err != nil {
		return Action{}, err
	}

	var a Action
	dec := json.NewDecoder(bytes.NewReader(obj))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&a); err != nil {
		return Action{}, fmt.Errorf("invalid JSON action: %w", err)
	}

	a.Type = strings.TrimSpace(a.Type)
	if a.Type == "" {
		return Action{}, fmt.Errorf("missing field: type")
	}

	switch a.Type {
	case ActionTypeTool:
		if strings.TrimSpace(a.Tool) == "" {
			return Action{}, fmt.Errorf("missing field: tool")
		}
		if len(a.Input) == 0 {
			a.Input = json.RawMessage([]byte(`{}`))
		}
	case ActionTypeDone:
		if strings.TrimSpace(a.Command) == "" {
			return Action{}, fmt.Errorf("missing field: command")
		}
	case ActionTypeAnswer:
		if strings.TrimSpace(a.Explanation) == "" {
			return Action{}, fmt.Errorf("missing field: explanation")
		}
	default:
		return Action{}, fmt.Errorf("unknown type: %s", a.Type)
	}

	return a, nil
}

func extractFirstJSONObject(text string) ([]byte, error) {
	s := strings.TrimSpace(text)
	// Common model behavior: wrap in ```json ... ``` fences.
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)

	start := strings.IndexByte(s, '{')
	if start < 0 {
		return nil, fmt.Errorf("no JSON object found")
	}

	inString := false
	escaped := false
	depth := 0
	for i := start; i < len(s); i++ {
		c := s[i]
		if escaped {
			escaped = false
			continue
		}
		if c == '\\' && inString {
			escaped = true
			continue
		}
		if c == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		switch c {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return []byte(s[start : i+1]), nil
			}
		}
	}

	return nil, fmt.Errorf("unterminated JSON object")
}
