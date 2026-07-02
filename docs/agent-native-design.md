# Agent-Native Design

NextVibe is designed for coding agents to call directly.

Human-readable Markdown is useful, but it is the fallback. The primary interface is a stable command protocol:

```bash
nextvibe scan --json
nextvibe suggest --json
nextvibe task --json
nextvibe check --json
```

## Division Of Responsibility

```text
Codex / Claude Code / Cursor = reasoning + editing + command execution
NextVibe = project state + next-step protocol + task navigation
```

NextVibe does not decide by free-form model reasoning. It inspects local files, applies deterministic rules, and returns structured output.

## Agent Contract

Agents should:

- run NextVibe before broad continuation work
- parse JSON rather than scrape Markdown
- respect allowed actions and forbidden changes
- keep the task scope narrow
- run `nextvibe check --json` after editing
- update project state only within the current task boundary

Agents should not:

- ask the human to copy generated prompts
- expand a task into unrelated refactors
- add model API keys for NextVibe
- treat Markdown output as the primary protocol

## Why JSON First

JSON gives agents a predictable contract:

- project signals are booleans
- stages have stable ids
- recommendations have stable ids and priorities
- task boundaries are arrays
- checks have pass/fail state and next actions

This makes NextVibe easy to call from Codex, Claude Code, Cursor, MCP tools, shell scripts, and CI workflows.
