# Release-Ready Project Example

This tiny example documents the files a project should have before relying on
NextVibe for release-oriented work:

- `README.md` with the product purpose
- local test command detected by `nextvibe scan`
- `.nextvibe/current-task.md` created by `nextvibe task`
- completion evidence in task markdown when a task needs proof beyond file
  existence

Suggested flow:

```bash
nextvibe init
nextvibe install all
nextvibe scan --json
nextvibe suggest --json
nextvibe task --json
nextvibe check --json
```

For MCP clients, use `nextvibe mcp --root /path/to/project` and call
`nextvibe_scan`, `nextvibe_suggest`, `nextvibe_task`, then `nextvibe_check`.
