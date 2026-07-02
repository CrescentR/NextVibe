# 004 - Add a Chinese README document

Task ID: 004
Status: complete
Title: Add a Chinese README document

## Background

The project is intended to be open source and useful to both English and Chinese
readers. The existing README is English-only, while the product roadmap and
development flow already contain Chinese documentation.

## Goal

Add a Chinese README that explains the project positioning, installation,
commands, agent workflow, platform support, and current scope without changing
CLI behavior.

## Allowed Files

- README.md
- README.zh-CN.md
- docs/development-log.md
- .nextvibe/current-task.md
- .nextvibe/tasks/004-add-chinese-readme-document.md

## Forbidden Changes

- do not change CLI behavior
- do not add runtime dependencies
- do not change release workflows
- do not rewrite unrelated documentation

## Acceptance Criteria

- `README.zh-CN.md` exists and is written in Chinese
- The Chinese README covers build, install, commands, agent workflow, and platform support
- The English README links to the Chinese README
- The development log records this task
- NextVibe check passes for the active task
