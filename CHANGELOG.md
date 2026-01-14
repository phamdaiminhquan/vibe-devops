# Changelog

## [Unreleased]

## [v0.3.8] - Interactive Step Extension

**Previous Version:** v0.3.7

### 🧠 UX Improvements
- **Interactive Timeout Handling**: When the Agent runs out of steps (max 10), Vibe now pauses and asks: "Do you want to give it 10 more steps?". This prevents complex tasks from failing abruptly.

## [v0.3.7] - Interactive Safety & Windows Support

**Previous Version:** v0.3.6

### 🛡️ Interactive Safety
- **Permission Requests**: Instead of blocking unknown system commands, Vibe now asks: "Agent wants to run UNSAFE command '...'. Allow? (y/N)". This gives you full control without blocking the agent's creativity.
- **Windows Whitelist**: Added `tasklist`, `Get-Process`, `Get-Service` to the safe list by default.

## [v0.3.6] - Intelligence & Performance Tuning

**Previous Version:** v0.3.5

### 🧠 Agent Intelligence
- **Increased Step Budget**: Default agent max steps increased from 5 to 10. Complex investigations (e.g., checking processes, ports, and logs together) no longer timeout prematurely.
- **Efficient Strategy**: Agent is now instructed to "batch" shell commands (e.g., `ps && netstat`) instead of running them one by one, saving step budget.

## [v0.3.5] - Actionable Error UX

**Previous Version:** v0.3.4

### 💎 UX Polish
- **Actionable API Key Error**: Now suggests the exact copy-paste command (`vibe config api-key ...`) to fix invalid credentials instantly.

## [v0.3.4] - Friendly Errors Update

**Previous Version:** v0.3.3

### 🐛 Improvements
- **Friendly API Key Error**: Vibe now catches 400/401 API errors and proactively suggests checking `.vibe.yaml` instead of dumping a raw stack trace.

## [v0.3.3] - System Safety Update

**Previous Version:** v0.3.2

Introduces `SafeShellTool` to prevent Agent hanging on system checks.

### 🛡️ New Features
- **Safe Shell Tool**: Vibe Agent can now execute whitelisted system commands (`ps`, `netstat`, `curl`, `df`, `free`, `uptime`, `whoami`) directly.
- **Improved Performance**: Replaces slow and dangerous `grep` filesystem scans for system status checks. "Check my backend process" is now instant.

## [v0.3.2] - The "Smart UX" Update

**Previous Version:** v0.3.1

Addresses user feedback on AI ambiguity and repetitive confirmations.

### ✨ UX Improvements
- **Auto-Execute Safe Commands**: Vibe now automatically runs read-only commands (`ls`, `find`, `grep`, `cat`, `pwd`, `whoami`, `date`) without asking for confirmation. This significantly speeds up investigation workflows.
- **Smarter AI Prompts**:
  - AI now asks for clarification if requests are ambiguous (e.g., "What is 'be'?").
  - AI provides friendlier status thoughts (`[VIBE] Checking folder structure...`).

## [v0.3.1] - UX Hotfix

**Previous Version:** v0.3.0

Improvements based on initial VPS feedback.

### ✨ UX Improvements
- **Agent Visibility**: Added `[VIBE] ⏳ Thinking...` status spinner.
- **Tool Feedback**: Now displaying tool usage action (e.g., `[VIBE] 🛠 Using tool: read_file`).
- **Friendly Output**: Replaced technical "Explanation:" header with user-friendly `[VIBE]` prefix.

### 🐛 Bug Fixes
- **FS Restriction**: Relaxed filesystem checks to allow accessing system paths (like `/home`, `/var/log`) instead of restricting to workspace root. This is critical for DevOps tasks.

## [v0.3.0] - The Foundation Update

**Previous Version:** v0.2.10

This MAJOR release upgrades the core architecture to **Hexagonal Architecture**, ensuring long-term maintainability and extensibility. It also acts as the baseline for the upcoming Plugin/Skill system.

### New Features

- **Dependency Auto-Check**: Vibe now proactively checks for essential tools (`git`, `docker`) on startup.
  - Displays a warning table if tools are missing.
  - Provides direct download links/install hints.
  - Non-blocking: You can still use the agent even if checks fail (though performance may be limited).

- **Smart Session Management**:
  - Sessions now track `created_at` and `updated_at` timestamps.
  - Added support for session `metadata` and `tags` (preparing for future history search).
  - Improved session persistence logic (jsonfile adapter).

### 🏗 Architecture & Refactoring

- **Hexagonal Architecture**:
  - Completely decoupled CLI (`cmd/`) from Business Logic (`internal/app/`).
  - Introduced `internal/ports` for strict interface boundaries.
  - New packages: `bootstrap` (DI), `dependency` (Registry).
  
- **Agent Service**:
  - Migrated logging to structured logging (`log/slog`).
  - Improved error handling and self-correction logic.

### Bug Fixes

- Fixed `run` command sometimes crashing on large output (Self-Healing Loop optimized).
- Fixed Session initialization not saving correct timestamp.
