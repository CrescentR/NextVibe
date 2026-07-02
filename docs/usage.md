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
