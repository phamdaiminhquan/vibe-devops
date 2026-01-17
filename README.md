# vibe-devops

An open-source AI terminal agent for automated DevOps tasks.

## Features

- **Natural Language to Shell**: Convert plain English requests into executable shell commands.
- **Multi-Provider AI**: Choose your AI brain - Gemini, OpenAI (GPT-4o), or Ollama (local/offline).
- **Smart Streaming**: Real-time streaming of AI explanations with clean output (no raw JSON).
- **System Diagnostics**: `vibe diagnose` checks disk, RAM, Docker, network, services with AI analysis (`--ai` flag).
- **Vietnamese Support**: Auto-detects Vietnamese input and warns if fonts are missing on Linux.
- **Context Providers**: Use `@file`, `@git`, `@system`, `@logs` to inject context into your requests.
- **Safety First**: Shows every command for confirmation before execution.
- **Smart Session**: Remembers context across runs with metadata (time, status) and simple context management.
- **Dependency Auto-Check**: Proactively warns if essential tools (Docker, Git) are missing.
- **Extensible**: Hexagonal Architecture with pluggable AI providers and tools.
- **Cross-Platform**: Works on Linux, macOS, and Windows.

## Installation

### One-Line Install (Linux & macOS)

You can install `vibe` with a single command. This will download the latest release and place it in `/usr/local/bin`.

**Prerequisites**: `curl` must be installed.

```bash
curl -sSL https://raw.githubusercontent.com/phamdaiminhquan/vibe-devops/main/install.sh | sh
```

### Build from Source

If you prefer, you can build from source:

```bash
# Clone the repository
git clone https://github.com/phamdaiminhquan/vibe-devops.git
cd vibe-devops

# Build the binary
go build -o vibe
```

## Usage

### 1. Initialize Vibe

First, navigate to your project directory and run `vibe init`. This will create a `.vibe.yaml` file.

```bash
# Initialize in the current directory
vibe init .
```

### 2. Configure Your API Key

Use the `config` command to select a provider and set your Gemini API key.

```bash
# Select provider
vibe config provider gemini

# Set API key (will validate and prompt you to choose a model)
vibe config api-key "YOUR_GEMINI_API_KEY"
```

### 3. Run Commands

Now you can make requests in natural language. Vibe will generate a shell command, ask for your confirmation, and then execute it.

```bash
vibe "list all running docker containers"
```

By default, `vibe run` uses **agent mode** (read-only tools like listing/reading files) and **self-heal** (can iterate after execution using the command output when troubleshooting).

To disable agent mode (simple single-shot command suggestion):

```bash
vibe --agent=false "list all running docker containers"
```

For troubleshooting-style questions, self-heal helps Vibe read command output and continue iterating until it can explain the root cause:

```bash
vibe "explain why service X is not running"
```

To disable self-heal:

```bash
vibe --self-heal=false "explain why service X is not running"
```

### 4. Use Context Providers

Inject relevant context directly into your request using `@mentions`:

```bash
# Include file content
vibe "@file main.go fix the bug on line 42" --agent

# Include git status
vibe "@git status commit these changes" --agent

# Include system info
vibe "@system os install docker" --agent

# Include log analysis (auto-highlights errors)
vibe "@logs app.log:100 what's wrong?" --agent
```

### 5. System Diagnostics

Run comprehensive health checks on your system:

```bash
# Basic diagnostics (disk, RAM, Docker, network, services)
vibe diagnose

# With AI analysis for issues
vibe diagnose --ai
```

### 6. Switch Models

To switch the configured model later:

```bash
# Interactive picker
vibe model

# Or set directly
vibe model gemini-1.5-pro

# List available models
vibe model --list
```

Agent mode can also persist a compact memory across runs (rolling summary + recent context). By default it uses both:
- project scope: `./.vibe/sessions/<session>.json`
- global scope: `~/.vibe/sessions/<session>.json`

Useful flags:

```bash
# Choose a session name (default: "default")
vibe --session fincap "why is fincap-api not running"

# Control scope: none|project|global|both
vibe --session-scope both "..."

# Control context size (approx chars + max lines for the recent tail)
vibe --context-budget 8000 --context-recent-lines 40 "..."

# Disable reading previous memory but still write updates
vibe --resume=false "..."

# Disable persistence entirely
vibe --no-session "..."
```

**Example Interaction:**

```
$ vibe run "show me the last 5 git commits with their author"

🤖 Calling AI to generate command...
ℹ️  Note: Currently, Vibe only interprets the command you send without any additional context.

Vibe suggests the following command:

git log -5 --pretty=format:"%h - %an, %ar : %s"

Do you want to execute it? (y/N) y

Executing command...
a1b2c3d - John Doe, 2 hours ago : feat: add awesome new feature
e4f5g6h - Jane Smith, 5 hours ago : fix: resolve bug in user auth
i7j8k9l - John Doe, 1 day ago : docs: update installation guide
m1n2o3p - Jane Smith, 2 days ago : refactor: improve database queries
q4r5s6t - John Doe, 2 days ago : chore: release version 0.1.0

✅ Command executed successfully.
```

## Contributing

Contributions are welcome! Please read our `CONTRIBUTING.md` file for our core principles and development guidelines.

## License

This project is licensed under the terms specified in the LICENSE file.
