# vibe-devops

An open-source AI terminal agent for automated DevOps tasks.

## 🚀 Features

- **Natural Language to Shell**: Convert plain English requests into executable shell commands.
- **AI-Powered**: Uses providers like Gemini to generate commands.
- **Safety First**: Shows every command for confirmation before execution.
- **Extensible**: Pluggable AI provider architecture.
- **Cross-Platform**: Works on Linux, macOS, and Windows.

## 🛠️ Installation

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

## 📖 Usage

### 1. Initialize Vibe

First, navigate to your project directory and run `vibe init`. This will create a `.vibe.yaml` file.

```bash
# Initialize in the current directory
vibe init .
```

### 2. Configure Your API Key

Use the `config` command to set your Gemini API key.

```bash
vibe config set gemini.apikey "YOUR_GEMINI_API_KEY"
```

### 3. Run Commands

Now you can make requests in natural language. Vibe will generate a shell command, ask for your confirmation, and then execute it.

```bash
vibe run "list all running docker containers"
```

**Example Interaction:**

```
$ vibe run "show me the last 5 git commits with their author"

🤖 Calling AI to generate command...

✨ Vibe suggests the following command:

  git log -5 --pretty=format:"%h - %an, %ar : %s"

Do you want to execute it? (y/N) y

🚀 Executing command...
a1b2c3d - John Doe, 2 hours ago : feat: add awesome new feature
e4f5g6h - Jane Smith, 5 hours ago : fix: resolve bug in user auth
i7j8k9l - John Doe, 1 day ago : docs: update installation guide
m1n2o3p - Jane Smith, 2 days ago : refactor: improve database queries
q4r5s6t - John Doe, 2 days ago : chore: release version 0.1.0

✅ Command executed successfully.
```

## 🤝 Contributing

Contributions are welcome! Please read our `CONTRIBUTING.md` file for our core principles and development guidelines.

## 📝 License

This project is licensed under the terms specified in the LICENSE file.
