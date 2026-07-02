# 005 - Add output language selection

Task ID: 005
Status: complete
Title: Add output language selection

## Background

NextVibe now has English and Chinese public documentation. The CLI should also
let users choose the language used for human-readable output while preserving
stable JSON for agents and automation.

## Goal

Add a minimal language selection option for CLI text output, supporting English
and Chinese first.

## Allowed Files

- internal/cli/cli.go
- tests/cli_smoke_test.go
- README.md
- README.zh-CN.md
- CLAUDE.md
- docs/usage.md
- docs/development-log.md
- .nextvibe/current-task.md
- .nextvibe/tasks/005-add-output-language-selection.md

## Forbidden Changes

- do not change JSON field names or machine-readable schema
- do not add runtime dependencies
- do not change scanner, detector, planner, checker, or installer behavior
- do not add languages beyond English and Chinese in this task

## Acceptance Criteria

- Text commands accept `--lang en` and `--lang zh`
- `--language` works as an alias for `--lang`
- Help output documents the language option
- JSON output remains machine-readable and backwards compatible
- Tests cover Chinese text output and invalid language handling
- README and usage docs explain language selection
- Development log records this task
- NextVibe check passes for the active task
