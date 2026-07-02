# Development Log

This log records the project flow as NextVibe evolves. Keep entries short,
task-oriented, and tied to concrete verification commands.

## 2026-07-02

### Task 001: Add a minimal deployment configuration

Goal: create a local deployment path for the CLI without cloud services.

Changed:

- Added `Dockerfile` for a multi-stage Go CLI build.
- Added `.dockerignore` to keep the Docker build context small.
- Added `docs/deployment.md` with local image build and run commands.

Verified:

```text
go test ./...
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" ./cmd/nextvibe
nextvibe check --json
```

Notes:

- Docker itself was not available in the local shell, so the equivalent Go build
  command was used to verify the Dockerfile's core compile step.

### Task 002: Add cross-platform open-source release readiness

Goal: make the open-source release path fit Windows, macOS, and Linux.

Changed:

- Added GitHub Actions CI for Windows, macOS, and Linux.
- Added tagged release workflow for `linux`, `darwin`, and `windows` archives.
- Added `.gitattributes` for cross-platform line ending and binary handling.
- Updated README, usage, release, roadmap, and deployment docs.

Verified:

```text
go test ./...
GOOS/GOARCH cross-build check for linux, darwin, and windows on amd64 and arm64
git diff --check
nextvibe check --json
```

Notes:

- Release packaging intentionally stays at GitHub archive artifacts for now.
- Package-manager distribution such as Homebrew, winget, apt, npm, or Scoop is
  deferred to a later focused task.

### Task 003: Record development progress and workflow

Goal: make the roadmap usable as a living development guide and record progress
as work happens.

Changed:

- Added this development log.
- Added a development progress recording section to `docs/development-roadmap.md`.
- Added a NextVibe task boundary for this workflow.

Verified:

```text
nextvibe scan --json
nextvibe suggest --json
nextvibe task --json
nextvibe check --json
git diff --check
```

Next:

- Keep adding short entries after each focused task is completed.

### Task 004: Add a Chinese README document

Goal: add a Chinese README for open-source readers while keeping the English
README as the primary default entry.

Changed:

- Added `README.zh-CN.md` with Chinese project positioning, build, install,
  commands, agent workflow, JSON output, rule system, scope, and docs links.
- Added a Chinese language link at the top of `README.md`.
- Added a NextVibe task boundary for this documentation task.

Verified:

```text
nextvibe scan --json
nextvibe suggest --json
nextvibe task --json
nextvibe check --json
git diff --check
```

Next:

- Keep the Chinese README in sync when public-facing commands or release flow change.

### Task 005: Add output language selection

Goal: let users choose English or Chinese for human-readable CLI output while
keeping JSON stable for agents and automation.

Changed:

- Added `--lang` and `--language` options for text output.
- Added Chinese labels for scan, suggest, task, check, init, and install text output.
- Kept JSON output field names and structure unchanged.
- Added CLI smoke tests for Chinese output, alias handling, invalid languages,
  and JSON stability with a language flag.
- Updated README, Chinese README, and usage docs.

Verified:

```text
go test ./...
go run ./cmd/nextvibe scan --lang zh
go run ./cmd/nextvibe suggest --language zh
go run ./cmd/nextvibe scan --json --lang zh
go build -o nextvibe.exe ./cmd/nextvibe
.\nextvibe.exe scan --lang zh
nextvibe check --json
git diff --check
```

Next:

- Keep JSON stable if future localized fields are added.

### Task 006: Detect local test commands

Goal: begin the Better Rules phase by exposing likely local test commands in
`scan` output.

Changed:

- Added `testCommands` to scan JSON output.
- Detected `go test ./...`, `npm test`, `mvn test`, and `gradle test` from
  common project files.
- Added test command display to English and Chinese text scan output.
- Added smoke tests for JSON command detection and Chinese text output.
- Updated README, Chinese README, usage docs, and roadmap notes.

Verified:

```text
go test ./...
go run ./cmd/nextvibe scan --json
go run ./cmd/nextvibe scan --lang zh
go build -o nextvibe.exe ./cmd/nextvibe
.\nextvibe.exe scan --json
.\nextvibe.exe scan --lang zh
.\nextvibe.exe check --json
git diff --check
```

Next:

- Keep the next Better Rules slice focused, likely API contract or database
  discovery improvements.

### Task 007: Polish README and verify the full usage flow

Goal: make the README feel ready for open-source readers and prove the documented
first-use workflow against the current CLI.

Changed:

- Reworked the English README with badges, icon-backed feature highlights,
  command tables, project map, platform support, and a complete local flow.
- Reworked the Chinese README with the same structure and visual treatment.
- Added an end-to-end local flow section to `docs/usage.md`.
- Fixed malformed emphasis in `CLAUDE.md` so agent instructions remain readable.
- Added a NextVibe task boundary for this documentation and verification task.

Verified:

```text
go build -o nextvibe.exe ./cmd/nextvibe
nextvibe init --json
nextvibe install all --json
nextvibe scan --json
nextvibe suggest --json
nextvibe task --json
nextvibe check --json
nextvibe scan --lang zh
```

Evidence:

- The flow was run in a clean temporary project at
  `C:\Users\Admin\AppData\Local\Temp\nextvibe-e2e-6116c635d18549aea5a5341f22df2538`.
- `scan --json` reported `testCommands: ["go test ./..."]`.
- `check --json` returned `passed: true`.

Next:

- Keep README changes aligned with future CLI behavior and release packaging.
