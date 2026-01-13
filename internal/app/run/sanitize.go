package run

import "strings"

func sanitizeAIResponse(response string) string {
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "`")
	response = strings.TrimSuffix(response, "`")
	response = strings.TrimPrefix(response, "shell")
	return strings.TrimSpace(response)
}
