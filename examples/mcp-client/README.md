# MCP Client Example

This example shows the minimal NextVibe MCP client configuration.

```json
{
  "mcpServers": {
    "nextvibe": {
      "command": "nextvibe",
      "args": ["mcp", "--root", "/absolute/path/to/project"]
    }
  }
}
```

Expected tools:

- `nextvibe_scan`
- `nextvibe_suggest`
- `nextvibe_task`
- `nextvibe_check`

Use the tools in that order when recovering project direction or continuing a
bounded development task.
