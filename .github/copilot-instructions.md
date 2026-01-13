# Copilot instructions for vibe-devops

## Big picture (where logic lives)
- `main.go` is the entrypoint; it injects build metadata into Cobra via `cmd.SetVersionInfo(...)`.
- `cmd/` is the CLI composition root (Cobra): argument parsing, prompts/printing, wiring.
- `internal/` is the Clean/Hex core:
  - `internal/domain/` small stable types (chat messages, command suggestion).
  - `internal/ports/` interfaces (outbound ports like `Provider`, `Executor`, `Tool`, `ConfigStore`).
  - `internal/app/` use-cases (`run`, `config`, `agent`).
  - `internal/adapters/` implementations (Gemini provider, local executor, vibeyaml config store, filesystem tools).
- `pkg/config/` owns the on-disk `.vibe.yaml` schema + IO (this is the “source of truth” for config format).
- `pkg/ai/` is mostly legacy/provider utilities; it is still used for Gemini model listing (`ai.GetGeminiModels`).

## Key flows
- **Run flow**: `cmd/run.go` loads `.vibe.yaml`, instantiates a provider, then:
  - default mode → `internal/app/run.Service.SuggestCommand` (single prompt → single command)
  - agent mode (`--agent`) → `internal/app/agent.Service` (tool loop → final command + explanation)
  - always prompts for confirmation before execution, then runs via `internal/adapters/executor/local`.
- **Config flow**: `cmd/config.go` uses `internal/app/config.Service` + `internal/adapters/configstore/vibeyaml` to mutate `.vibe.yaml`; for Gemini it validates key + prompts a model using `pkg/ai.GetGeminiModels`.

## Agent mode protocol (important)
- Agent mode is opt-in: `vibe run --agent --agent-max-steps 5 "..."` (see `cmd/run.go`).
- The model must output EXACTLY one JSON object per step (see `internal/app/agent/protocol.go`):
  - tool call: `{ "type": "tool", "tool": "read_file", "input": { ... } }`
  - final: `{ "type": "done", "command": "...", "explanation": "..." }`
- Tools are read-only by contract (`internal/ports/tool.go`); examples: `internal/adapters/tools/fs/*`.

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
