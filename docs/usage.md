# Usage

## Build Locally

```bash
go build ./cmd/nextvibe
```

On Windows:

```powershell
.\nextvibe.exe --help
```

On macOS or Linux:

```bash
./nextvibe --help
```

## Install From A Release

Download the archive that matches your platform:

```text
nextvibe_linux_amd64.tar.gz
nextvibe_linux_arm64.tar.gz
nextvibe_darwin_amd64.tar.gz
nextvibe_darwin_arm64.tar.gz
nextvibe_windows_amd64.zip
nextvibe_windows_arm64.zip
```

Extract the archive and place the binary on your `PATH`.

On Windows the binary is `nextvibe.exe`:

```powershell
nextvibe.exe --help
```

On macOS or Linux the binary is `nextvibe`:

```bash
nextvibe --help
```

## Cross-Compile Locally

Go can build the CLI for other operating systems without adding runtime
dependencies.

macOS or Linux:

```bash
mkdir -p dist
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/nextvibe-linux-amd64 ./cmd/nextvibe
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o dist/nextvibe-darwin-arm64 ./cmd/nextvibe
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/nextvibe-windows-amd64.exe ./cmd/nextvibe
```

PowerShell:

```powershell
New-Item -ItemType Directory -Force -Path dist | Out-Null
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"; $env:GOARCH = "amd64"; go build -trimpath -ldflags="-s -w" -o dist/nextvibe-linux-amd64 ./cmd/nextvibe
$env:GOOS = "darwin"; $env:GOARCH = "arm64"; go build -trimpath -ldflags="-s -w" -o dist/nextvibe-darwin-arm64 ./cmd/nextvibe
$env:GOOS = "windows"; $env:GOARCH = "amd64"; go build -trimpath -ldflags="-s -w" -o dist/nextvibe-windows-amd64.exe ./cmd/nextvibe
```

## Initialize A Repo

```bash
nextvibe init
```

This creates `.nextvibe/` and only fills in missing files. Existing project
state is preserved. New workspaces include `.nextvibe/config.yaml` and a starter
rule template at `.nextvibe/rules/default-task.md`.

## End-To-End Local Flow

From a project root, run the full local workflow:

```bash
nextvibe init
nextvibe install all
nextvibe scan --json
nextvibe suggest --json
nextvibe task --json
nextvibe check --json
```

This verifies the complete first-use path:

- `.nextvibe/` is created or updated.
- Agent integration files are installed without replacing user-authored content.
- `scan --json` returns project signals, risks, stacks, and likely test commands.
- `suggest --json` returns one recommended next task.
- `task --json` creates or reads the active task.
- `check --json` reports whether the active task's completion evidence passes.

For human-readable Chinese output:

```bash
nextvibe scan --lang zh
nextvibe suggest --lang zh
nextvibe check --lang zh
```

## Install Agent Integrations

```bash
nextvibe install codex
nextvibe install claude
nextvibe install cursor
nextvibe install all
```

Generated integration files teach agents to call:

```bash
nextvibe scan --json
nextvibe suggest --json
nextvibe task --json
nextvibe check --json
```

Verify installed integrations:

```bash
nextvibe install verify --json
```

## Run As An MCP Server

```bash
nextvibe mcp
nextvibe mcp --root /path/to/project
```

MCP clients can call:

- `nextvibe_scan`
- `nextvibe_suggest`
- `nextvibe_task`
- `nextvibe_check`

See [MCP](mcp.md) for client configuration.

## Output Language

Human-readable text output supports English and Chinese:

```bash
nextvibe scan --lang en
nextvibe scan --lang zh
nextvibe suggest --language zh
nextvibe check --lang zh
```

`--language` is an alias for `--lang`. JSON output keeps stable field names and
structure, so agents can keep using commands such as:

```bash
nextvibe scan --json --lang zh
```

## Daily Agent Loop

When the user asks to continue development, the agent should run:

```bash
nextvibe scan --json
nextvibe suggest --json
nextvibe task --json
```

The agent then edits files within the returned task boundaries.

After editing:

```bash
nextvibe check --json
```

The check command verifies completion evidence for the active task. Older tasks
still use allowed-file checks. Newer task files can add:

```markdown
## Required Commands

- go test ./...

## Evidence Files

- docs/api/openapi.yaml
```

`check --json` runs required commands, checks evidence files or directories, and
still verifies changed files stay inside the active task boundary.

## Package Manager Install Surfaces

Distribution scaffolding is available for:

- npm: `package.json`, `npm/nextvibe.js`, `npm/postinstall.js`
- Homebrew: `Formula/nextvibe.rb`
- Scoop: `scoop/nextvibe.json`

See [Distribution](distribution.md) and [Release Checklist](release-checklist.md)
for the release-time checksum replacement points and quality gate.

## Test Command Detection

`scan` reports likely local test commands when it can infer them from common
project files:

```bash
nextvibe scan
nextvibe scan --json
```

Currently detected commands include:

- `go test ./...` for Go modules
- `npm test` for Node projects with a `test` script
- `mvn test` for Maven projects
- `gradle test` for Gradle projects
