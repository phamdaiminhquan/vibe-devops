package run

import "fmt"

func buildPrompt(goos, userRequest string) string {
	return fmt.Sprintf(
		`You are an expert AI assistant specializing in shell commands. Your task is to convert a user's request into a single, executable shell command for a %s environment.
- Only output the raw command.
- Do not include any explanation, markdown, backticks, or any text other than the command itself.
- If the request is ambiguous or unsafe, reply with "Error: Ambiguous or unsafe request."

User's request: "%s"
Shell command:`,
		goos,
		userRequest,
	)
}
