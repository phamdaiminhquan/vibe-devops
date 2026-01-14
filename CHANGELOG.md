# Changelog

## [Unreleased]

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
