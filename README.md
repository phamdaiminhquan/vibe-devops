# vibe-devops

An open-source AI terminal agent for automated VPS management and self-healing Docker deployments.

## 🚀 Features

- **AI-Powered DevOps**: Leverage AI to automate VPS and Docker management tasks
- **CLI Interface**: Easy-to-use command-line interface built with Cobra
- **Extensible**: Modular architecture with pluggable AI providers
- **Shell Command Execution**: Run shell commands with AI assistance

## 📋 Prerequisites

- Go 1.24 or higher

## 🛠️ Installation

### Build from Source

```bash
git clone https://github.com/phamdaiminhquan/vibe-devops.git
cd vibe-devops
go build -o vibe-devops
```

### Install

```bash
go install github.com/phamdaiminhquan/vibe-devops@latest
```

## 📖 Usage

### Initialize a Project

Scan a directory to initialize vibe-devops configuration:

```bash
vibe-devops init [directory]
```

Example:
```bash
# Initialize in current directory
vibe-devops init .

# Initialize in specific directory
vibe-devops init /path/to/project
```

### Run Shell Commands

Execute shell commands with AI-powered assistance:

```bash
vibe-devops run [command]
```

Examples:
```bash
# Run a simple command
vibe-devops run ls -la

# Run with verbose output
vibe-devops run -v "docker ps"

# Execute complex commands
vibe-devops run "echo 'Hello from vibe-devops!'"
```

### Get Help

```bash
# Show main help
vibe-devops --help

# Show version
vibe-devops --version

# Show help for specific command
vibe-devops init --help
vibe-devops run --help
```

## 📁 Project Structure

```
vibe-devops/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command setup
│   ├── init.go            # Init command (directory scanning)
│   └── run.go             # Run command (shell execution)
├── pkg/                    # Packages
│   └── ai/                # AI provider interfaces
│       ├── provider.go    # AI provider interface definitions
│       └── mock.go        # Mock provider implementation
└── main.go                # Application entry point
```

## 🔧 Development

### Build

```bash
go build -o vibe-devops
```

### Run Tests

```bash
go test ./...
```

### Add Dependencies

```bash
go get <package>
go mod tidy
```

## 🤖 AI Provider Interface

The project includes a flexible AI provider interface that allows integration with various AI services:

```go
type Provider interface {
    GetCompletion(prompt string) (string, error)
    GetName() string
    IsConfigured() bool
}
```

### Available Providers

- **Mock Provider**: A simple mock implementation for testing

### Adding New Providers

1. Implement the `Provider` interface in `pkg/ai/`
2. Add configuration support
3. Register the provider in your application

## 📝 License

This project is licensed under the terms specified in the LICENSE file.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📧 Contact

For questions or feedback, please open an issue on GitHub.
