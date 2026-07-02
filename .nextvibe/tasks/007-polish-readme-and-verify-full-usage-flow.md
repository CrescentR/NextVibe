# 007 - Polish README and verify the full usage flow

Task ID: 007
Status: complete
Title: Polish README and verify the full usage flow

## Background

The project is moving toward open-source readiness. The README should feel more
polished and easier to scan, and the documented first-use flow should be proven
against the current CLI.

## Goal

Refresh the English and Chinese README documents with a more polished
presentation, then verify the full local usage flow from build through
`init/install/scan/suggest/task/check`.

## Allowed Files

- README.md
- README.zh-CN.md
- docs/usage.md
- docs/roadmap.md
- docs/development-log.md
- .nextvibe/current-task.md
- .nextvibe/tasks/005-add-output-language-selection.md
- .nextvibe/tasks/006-detect-local-test-commands.md
- .nextvibe/tasks/007-polish-readme-and-verify-full-usage-flow.md
- CLAUDE.md
- internal/cli/cli.go
- internal/detector/detector.go
- tests/cli_smoke_test.go

## Forbidden Changes

- do not change CLI behavior for this task
- do not add runtime dependencies
- do not add website or landing-page code
- do not add package-manager distribution in this task

## Acceptance Criteria

- English README has a polished open-source layout with visual badges or icons
- Chinese README receives the same information architecture and visual treatment
- README documents a complete first-use flow from install/build through `check`
- Usage docs include the verified end-to-end flow
- The current CLI successfully runs the full flow in a clean sample project
- Development log records this task and verification evidence
- `go test ./...`, `git diff --check`, and `nextvibe check --json` pass
