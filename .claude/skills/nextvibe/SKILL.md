<!-- NEXTVIBE:START -->
# NextVibe Project Navigation

Use this skill when the user asks to continue the project, recover direction, decide the next development step, or plan what to do after an AI-assisted build stalls.

Workflow:

1. Run `nextvibe scan --json` to inspect project state.
2. Run `nextvibe suggest --json` to get the next recommended task.
3. Run `nextvibe task --json` to create or read the active task.
4. Follow the task boundaries exactly.
5. After edits, run `nextvibe check --json`.

Do not ask the user to copy a generated prompt into Claude Code. Call the local CLI directly and use the JSON result.
<!-- NEXTVIBE:END -->
