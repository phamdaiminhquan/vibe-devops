# Contributing to Vibe-DevOps

First off, thank you for considering contributing! This project is an open-source AI terminal agent designed for clarity, extensibility, and ease of contribution. To maintain these goals, we adhere to a few core principles.

## Core Principles ("Hard Laws")

These are the fundamental rules that guide our architecture and development. They ensure that `vibe` remains a high-quality, maintainable, and powerful tool.

### 1. Hexagonal Architecture (Clean Arch)

We follow Hexagonal Architecture to separate checks and balances:
- **`cmd/`**: CLI Interface (Cobra). Minimal logic.
- **`internal/ports/`**: Interfaces (Provider, SessionStore, etc.).
- **`internal/adapters/`**: Real implementations (Gemini, FS, JSON Store).
- **`internal/app/`**: Business Logic (Use Cases).

### 2. The Provider Pattern is Absolute

All AI logic **must** be implemented through the `Provider` interface defined in `internal/ports/provider.go`. This is the most critical rule.

- **Why?** It decouples the core application from any specific AI service (Gemini, OpenAI, etc.). This allows users to easily switch providers and enables developers to add new providers without touching the core command logic.
- **How?** To add a new AI, create a new adapter (e.g., `internal/adapters/provider/openai/`) and implement the `Provider` interface. Then, wire it in `internal/app/bootstrap`.

### 3. Configuration is King

All user-specific settings, especially secrets like API keys, **must** be managed through a configuration file (`.vibe.yaml`).

- **Why?** It avoids hardcoding secrets and allows users to easily configure the agent for their specific needs. It also makes the application environment-agnostic.
- **How?** The `init` command should generate a default, well-commented configuration file. The application should load this configuration at startup via `pkg/config`.

### 4. Commands are for Humans, Logic is for Handlers

The `cmd/` directory is strictly for defining the user-facing CLI commands and handling their flags and arguments. All business logic **must** reside in `internal/app/command/`.

- **Example:** The `run` command in `cmd/run.go` should parse the user's request, initialize the app via `bootstrap`, and delegate to `RunHandler.Handle()` in `internal/app/command/run_handler.go`.

### 5. Interfaces Define the Future

Use Go interfaces to define the boundaries between different parts of the application. The AI `Provider`, `SessionStore`, and `Logger` are examples.

- **Why?** Interfaces are the key to extensibility. They allow components to be swapped out and tested in isolation.

## How to Contribute

1.  **Fork the repository.**
2.  **Create a new branch** for your feature or bug fix.
3.  **Write your code,** adhering to the Core Principles above.
4.  **Add tests** for your changes.
5.  **Ensure your code is formatted** with `gofmt`.
6.  **Submit a Pull Request** with a clear description of your changes.

Thank you for helping us build the future of DevOps!
