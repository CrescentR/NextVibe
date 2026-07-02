# 006 - Detect local test commands

Task ID: 006
Status: complete
Title: Detect local test commands

## Background

The development roadmap's next phase is Better Rules. One focused rule
improvement is language-specific test command detection, so agents can see how a
project should be verified locally before editing or after completing a task.

## Goal

Expose likely local test commands in `scan` output for common project stacks,
starting with Go, Node package scripts, Maven, and Gradle.

## Allowed Files

- internal/detector/detector.go
- internal/cli/cli.go
- tests/cli_smoke_test.go
- README.md
- README.zh-CN.md
- docs/usage.md
- docs/roadmap.md
- docs/development-log.md
- .nextvibe/current-task.md
- .nextvibe/tasks/005-add-output-language-selection.md
- .nextvibe/tasks/006-detect-local-test-commands.md
- CLAUDE.md

## Forbidden Changes

- do not implement configurable rules yet
- do not add runtime dependencies
- do not change task generation or checker behavior
- do not add MCP work in this task

## Acceptance Criteria

- `scan --json` includes detected local test commands
- Text `scan` output shows detected test commands in English and Chinese modes
- Go projects detect `go test ./...`
- Node projects with a `test` script detect `npm test`
- Maven projects detect `mvn test`
- Gradle projects detect `gradle test`
- Tests cover command detection and JSON shape
- README and usage docs mention test command detection
- Development log records this task
- NextVibe check passes for the active task
