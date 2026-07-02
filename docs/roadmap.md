# Roadmap

## Phase 1: Local CLI

Status: current.

- initialize `.nextvibe/`
- scan project structure
- detect stacks and project stage
- suggest a next task with deterministic rules
- create or read the active task
- check basic task completion
- install Codex, Claude Code, and Cursor integration files
- return stable JSON for agent consumption
- run CI on Windows, macOS, and Linux
- publish basic cross-platform release archives from version tags

## Phase 2: Better Rules

- richer framework detection
- language-specific test command detection
- API contract discovery across more file names
- database and migration detection by ecosystem
- task templates by project stage
- configurable rule overrides in `.nextvibe/`

## Phase 3: Agent Integrations

- stronger Codex skill support
- richer Claude Code skill and command workflows
- Cursor rule variants
- generated agent context summaries
- explicit task boundary updates after successful checks

## Phase 4: Protocol Surface

- MCP server exposing the same scan, suggest, task, and check primitives
- schema versioning for JSON output
- machine-readable task history
- optional editor and CI integrations

MCP is a tool exposure layer, not a model calling layer. NextVibe should remain local-first and model-agnostic.
