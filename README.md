# NextVibe

NextVibe is an agent-native project navigation tool for AI-assisted developers.

It does not replace Codex, Claude Code, or Cursor. It gives them a reliable local command they can call to understand project state, choose a bounded next task, and verify whether that task is complete.

No model API key. No extra chat UI. No copy-paste prompt workflow.

Install it once, then let your coding agent call it like any other development tool.

## Why It Exists

AI coding tools can quickly create a prototype, scaffold a repo, or generate a frontend. Many projects then stall because the next step is unclear: API contract, backend boundary, database model, tests, deployment, or more UI.

NextVibe answers one narrow question:

> What should the coding agent build next?

It does this with local project scanning, rule-based stage detection, stable JSON output, and task boundary files under `.nextvibe/`.

## Product Principles

- NextVibe does not call OpenAI, Anthropic, or any other model API.
- NextVibe does not require users to configure model keys.
- NextVibe is not another AI chat application.
- NextVibe is not a prompt generator that asks humans to copy text into an agent.
- NextVibe is a CLI plus agent integration layer with machine-readable output.
- Markdown files are project state and fallback documentation, not the primary interaction model.

The intended flow is:

```text
agent reads instructions
agent calls nextvibe command
agent parses JSON
agent edits code
agent calls nextvibe check
agent updates project state
```

Not:

```text
human runs nextvibe
human copies prompt
human pastes prompt into agent
```

## Build

Requires Go 1.22 or newer.

```bash
go build ./cmd/nextvibe
```

On macOS or Linux this creates `./nextvibe`.

On Windows this creates `nextvibe.exe`:

```powershell
.\nextvibe.exe scan --json
```

## Platform Support

NextVibe is intended to run on Windows, macOS, and Linux.

Continuous integration runs the test suite on all three operating systems and
cross-compiles release binaries for:

- `linux/amd64`
- `linux/arm64`
- `darwin/amd64`
- `darwin/arm64`
- `windows/amd64`
- `windows/arm64`

Tagged GitHub releases publish `.tar.gz` archives for macOS and Linux and `.zip`
archives for Windows. See [docs/release.md](docs/release.md) for the maintainer
release flow.

## Core Commands

```bash
nextvibe init
nextvibe scan
nextvibe suggest
nextvibe task
nextvibe check
nextvibe install codex
nextvibe install claude
nextvibe install cursor
nextvibe install all
```

Agent-facing commands support stable JSON output:

```bash
nextvibe scan --json
nextvibe suggest --json
nextvibe task --json
nextvibe check --json
```

`init` and `install` also support `--json` for automation.

## First-Time Setup In A Project

From the target project root:

```bash
nextvibe init
nextvibe install all
```

This creates or updates:

```text
.nextvibe/
  project.md
  stage.md
  roadmap.md
  current-task.md
  decisions.md
  agent-context.md
  tasks/

AGENTS.md
CLAUDE.md
.cursor/rules/nextvibe.mdc
.claude/skills/nextvibe/SKILL.md
.claude/commands/nv-suggest.md
.claude/commands/nv-check.md
```

Existing files are preserved. NextVibe appends or updates a marked section instead of replacing user-authored instructions.

## Install From A Release

Download the archive for your operating system from the GitHub release page,
extract it, and put the `nextvibe` binary on your `PATH`.

On Windows the binary is named `nextvibe.exe`. On macOS and Linux it is named
`nextvibe`.

## Agent Workflow

When a user says "continue this project" or "what should I do next?", the coding agent should run:

```bash
nextvibe scan --json
nextvibe suggest --json
nextvibe task --json
```

The agent should then use the returned task boundaries:

- allowed files or actions
- forbidden changes
- acceptance criteria
- expected artifacts

After editing, the agent should run:

```bash
nextvibe check --json
```

## JSON Shape

`nextvibe scan --json` returns project state:

```json
{
  "projectName": "example-project",
  "detectedStacks": ["go", "react"],
  "keyFiles": ["README.md", "go.mod", "package.json"],
  "keyDirectories": ["components", "mock", "src"],
  "signals": {
    "hasFrontend": true,
    "hasBackend": true,
    "hasMockData": true,
    "hasApiContract": false,
    "hasDatabaseSchema": false,
    "hasTests": false,
    "hasDockerfile": false,
    "hasAgentInstructions": false
  },
  "stage": {
    "id": "frontend-prototype-backend-incomplete",
    "label": "Frontend prototype exists, backend contract missing"
  },
  "risks": ["Mock data detected", "No API contract found"]
}
```

`nextvibe suggest --json` returns a recommended next task. `nextvibe task --json` creates or reads the active task. `nextvibe check --json` verifies basic completion signals.

## Rule System

The first version is intentionally local and deterministic. It uses visible repository signals such as:

- `package.json`, `go.mod`, `pom.xml`, `build.gradle`
- `README.md`, `AGENTS.md`, `CLAUDE.md`
- `src/`, `app/`, `pages/`, `components/`
- `api/`, `server/`, `backend/`, `internal/`, `cmd/`
- `mock/`, `mocks/`
- `docs/api/`, `openapi.yaml`
- `migrations/`, `sql/`
- `test/`, `tests/`, `__tests__`, and common test filename suffixes
- `Dockerfile`, `docker-compose.yml`

Priority rules start simple:

1. Missing project goal or implementation structure: write project goals first.
2. Frontend exists but API contract is missing: design the API contract.
3. API contract exists but data model is missing: design the data model.
4. Backend or business code exists but tests are missing: add focused tests.
5. Deployment config is missing: add a minimal deployment path.
6. Agent integration files are missing: run `nextvibe install all`.

## Naming And Rebranding

The project name, CLI command name, and workspace directory are centralized in:

```text
internal/brand/brand.go
```

If the project is renamed later, update those constants first, then regenerate docs and integration files.

## Current Scope

In scope for the first version:

- cross-platform Go CLI
- `.nextvibe/` project state directory
- project scanning
- rule-based stack and stage detection
- next-step recommendation
- current task generation
- basic task checking
- Codex, Claude Code, and Cursor integration files
- stable JSON output for agents
- cross-platform CI and release archives for Windows, macOS, and Linux

Out of scope for now:

- web UI
- cloud service
- account system
- model API integration
- MCP server
- prompt marketplace

MCP can be added later as another way to expose the same local protocol. It should not change the core principle: NextVibe helps agents navigate local project state; it does not become the model.
