# Contributing to Vibe-DevOps

First off, thank you for considering contributing! This project is an open-source AI terminal agent designed for clarity, extensibility, and ease of contribution. To maintain these goals, we adhere to a few core principles.

## Core Principles ("Hard Laws")

These are the fundamental rules that guide our architecture and development. They ensure that `vibe` remains a high-quality, maintainable, and powerful tool.

### 1. The Provider Pattern is Absolute

All AI logic **must** be implemented through the `Provider` interface defined in `pkg/ai/provider.go`. This is the most critical rule.

- **Why?** It decouples the core application from any specific AI service (Gemini, OpenAI, etc.). This allows users to easily switch providers and enables developers to add new providers without touching the core command logic.
- **How?** To add a new AI, create a new file (e.g., `pkg/ai/openai.go`) and implement the `Provider` interface. Then, register it in the application's provider factory.

### 2. Configuration is King

All user-specific settings, especially secrets like API keys, **must** be managed through a configuration file (`.vibe.yaml`).

- **Why?** It avoids hardcoding secrets and allows users to easily configure the agent for their specific needs. It also makes the application environment-agnostic.
- **How?** The `init` command should generate a default, well-commented configuration file. The application should load this configuration at startup.

### 3. Commands are for Humans, Logic is for Packages

The `cmd/` directory is strictly for defining the user-facing CLI commands and handling their flags and arguments. All business logic **must** reside in the `pkg/` directory.

- **Why?** This is a standard Go project layout that promotes separation of concerns. It makes the code easier to read, test, and reuse. The CLI is just a user interface; the core logic is what matters.
- **Example:** The `run` command in `cmd/run.go` should parse the user's request. It then calls a function in a package (e.g., `pkg/runner/runner.go`) which is responsible for coordinating with the AI provider, executing the command, and returning the result.

### 4. Interfaces Define the Future

Use Go interfaces to define the boundaries between different parts of the application. The AI `Provider` is one example. This could also apply to future concepts like `Executor` (for running commands locally vs. SSH) or `Logger`.

- **Why?** Interfaces are the key to extensibility. They allow components to be swapped out and tested in isolation. They are the blueprint for what the application can become.

## How to Contribute

1.  **Fork the repository.**
2.  **Create a new branch** for your feature or bug fix.
3.  **Write your code,** adhering to the Core Principles above.
4.  **Add tests** for your changes.
5.  **Ensure your code is formatted** with `gofmt`.
6.  **Submit a Pull Request** with a clear description of your changes.

Thank you for helping us build the future of DevOps!
