# Copilot instructions for vibe-devops

## Big picture (where logic lives)
- `main.go` is the entrypoint; it injects build metadata into Cobra via `cmd.SetVersionInfo(...)`.
- `cmd/` is the CLI composition root (Cobra): argument parsing, prompts/printing, wiring.
- `internal/` is the Clean/Hex core:
  - `internal/domain/` small stable types (chat messages, command suggestion).
  - `internal/ports/` interfaces (outbound ports like `Provider`, `Executor`, `Tool`, `ConfigStore`).
  - `internal/app/` use-cases:
    - `bootstrap`: app initialization and wiring (DI container).
    - `command`: handlers for CLI commands (e.g. `RunHandler`).
    - `agent`, `run`, `config`, `session`, `dependency`.
  - `internal/adapters/` implementations (Gemini provider, local executor, vibeyaml config store, filesystem tools).
- `pkg/config/` owns the on-disk `.vibe.yaml` schema + IO (this is the “source of truth” for config format).
- `pkg/ai/` is mostly legacy/provider utilities; it is still used for Gemini model listing (`ai.GetGeminiModels`).

## Key flows
- **Run flow**: `cmd/run.go` delegates completely to `bootstrap` (init) and `internal/app/command/run_handler.go` (logic):
  - 1. `bootstrap.Initialize()`: loads config, sets up providers and logger.
  - 2. `dependency.Manager.VerifyAll()`: checks if Git/Docker are installed.
  - 3. delegates to `agent.Service` (loop) or `run.Service` (single-shot).
  - 4. `run_handler` manages the self-heal loop and confirmation UI.
  - always prompts for confirmation before execution, then runs via `internal/adapters/executor/local`.

- Convenience: `vibe "..."` is treated as `vibe run "..."` (see `cmd/root.go`).
- **Config flow**: `cmd/config.go` uses `internal/app/config.Service` + `internal/adapters/configstore/vibeyaml` to mutate `.vibe.yaml`; for Gemini it validates key + prompts a model using `pkg/ai.GetGeminiModels`.

## Agent mode protocol (important)
- Agent mode is the default for `vibe run` (see `cmd/run.go`). Disable with `--agent=false`.
- The model must output EXACTLY one JSON object per step (see `internal/app/agent/protocol.go`):
  - tool call: `{ "type": "tool", "tool": "read_file", "input": { ... } }`
  - final: `{ "type": "done", "command": "...", "explanation": "..." }`
- For post-execution analysis (self-heal loop), the model can also return:
  - answer: `{ "type": "answer", "explanation": "..." }`
- Tools are read-only by contract (`internal/ports/tool.go`); examples: `internal/adapters/tools/fs/*`.

## Session memory (project + global)
- Agent mode can persist a compact rolling summary + recent tail for context reuse across runs:
  - Project: `./.vibe/sessions/<name>.json`
  - Global: `~/.vibe/sessions/<name>.json`
- Implementation: `internal/app/session` + `internal/adapters/sessionstore/jsonfile`.

## Configuration format
- File name: `.vibe.yaml` (see `pkg/config/config.go`). Schema:
  - `ai.provider` (currently `gemini`)
  - `ai.gemini.apiKey`, `ai.gemini.model`
- Placeholder key `YOUR_GEMINI_API_KEY_HERE` means “not configured”.

## Adding a new provider/tool (project-specific pattern)
- New AI provider: implement `internal/ports.Provider`, add adapter in `internal/adapters/provider/<name>/`, then wire selection in `cmd/run.go` (and `cmd/config.go` if it needs setup/validation).
- New agent tool: implement `internal/ports.Tool`, keep it deterministic + side-effect free, then register it in `cmd/run.go` under agent mode.

## Developer workflows
- Build: `go build ./...`  |  Test: `go test ./...`  |  CLI: `go run . --help`
- Release: GoReleaser (`.goreleaser.yml`) via `.github/workflows/release.yml` on tags `v*` (injects `main.version`, `main.commit`, `main.date`).
- When changing `.vibe.yaml` schema, update `pkg/config/config.go`, `cmd/init.go`, `cmd/config.go`, and `README.md`.
