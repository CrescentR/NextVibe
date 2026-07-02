# Vision

NextVibe helps AI coding agents keep projects moving after the first burst of generation.

The product is built around a simple idea: coding agents are already good at reasoning, editing files, and running commands. What they often lack is durable project navigation state. NextVibe supplies that state through a local CLI and agent-readable JSON.

## Positioning

NextVibe tells your coding agent what to build next.

Slogan:

```text
From vibe coding to vibe shipping.
```

NextVibe is for:

- AI programming users
- vibe coders
- independent developers
- small teams using Codex, Claude Code, Cursor, or similar tools

It is not another coding agent. It is the local project navigation tool those agents can call.

## What It Optimizes For

- deterministic local behavior
- readable project state
- stable JSON contracts
- narrow task boundaries
- explicit acceptance criteria
- safe integration with existing agent instruction systems

## Non-Goals

- no model API calls
- no model key configuration
- no chat UI
- no copy-paste prompt loop
- no cloud dependency

The first version should feel like `git`, `npm`, `go test`, or `python`: a normal local development command that an agent can use without ceremony.
