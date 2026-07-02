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

This creates `.nextvibe/` and only fills in missing files. Existing project state is preserved.

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

The check command verifies basic completion signals for the active task.
