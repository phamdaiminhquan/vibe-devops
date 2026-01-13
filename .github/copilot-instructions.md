# Copilot instructions for vibe-devops

## Big picture
- `main.go` boots the CLI and injects build metadata (version/commit/date) into Cobra.
- `cmd/` contains Cobra commands only (argument parsing, prompts, printing, wiring).
- `pkg/` contains reusable logic:
  - `pkg/ai/` implements the provider pattern (`Provider` interface in `pkg/ai/provider.go`).
  - `pkg/config/` owns the on-disk config format and IO for `.vibe.yaml`.

## Key flows
- **Run flow**: `cmd/run.go` loads `.vibe.yaml` via `config.Load(".")`, selects a provider by name, builds a prompt for the current OS, shows the generated command, then executes it only after confirmation.
- **Config flow**: `cmd/config.go` updates `.vibe.yaml` in-place using `config.Write(".", cfg)` and (for Gemini) validates the API key by calling `ai.GetGeminiModels` and prompting the user to pick a model.

## Configuration format
- Config file name is `.vibe.yaml` (see `pkg/config/config.go`).
- Current schema:
  - `ai.provider`: provider string (currently `gemini`)
  - `ai.gemini.apiKey`, `ai.gemini.model`
- Default placeholder key is `YOUR_GEMINI_API_KEY_HERE`; treat it as “not configured”.

## Provider pattern (repo rule)
- All new AI integrations must implement `Provider` (see `pkg/ai/provider.go`).
- Add a provider as a new file in `pkg/ai/` (example: `pkg/ai/gemini.go`) and wire selection in `cmd/run.go` based on `cfg.AI.Provider`.

## Developer workflows
- Build: `go build ./...`
- Test: `go test ./...`
- Local CLI usage: `go run . --help`
- Release automation:
  - CI: `.github/workflows/release.yml` runs GoReleaser on tags `v*`.
  - GoReleaser config: `.goreleaser.yml` builds `linux/windows/darwin` for `amd64/arm64` with `CGO_ENABLED=0` and injects `main.version`, `main.commit`, `main.date`.
  - Install script: `install.sh` fetches the latest GitHub Release asset for `linux/darwin` and installs to `/usr/local/bin`.

## Conventions worth keeping
- Prefer small, explicit wiring in `cmd/` and keep provider/config logic in `pkg/` (see `CONTRIBUTING.md` “Commands are for Humans, Logic is for Packages”).
- When changing `.vibe.yaml` schema, update:
  - `pkg/config/config.go` (types + defaults)
  - `cmd/init.go` (generated config)
  - `cmd/config.go` (mutations/validation)
  - `README.md` usage snippets
