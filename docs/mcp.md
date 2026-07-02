# MCP Server

NextVibe exposes its local project navigation protocol through a stdio MCP
server:

```bash
nextvibe mcp
```

The server reads newline-delimited JSON-RPC messages from stdin and writes
responses to stdout. It does not call model APIs and does not require network
access.

## Client Configuration

Point an MCP client at the project root you want NextVibe to inspect:

```json
{
  "mcpServers": {
    "nextvibe": {
      "command": "nextvibe",
      "args": ["mcp", "--root", "/path/to/project"]
    }
  }
}
```

If `--root` is omitted, NextVibe uses the process working directory.

## Tools

The server exposes four tools:

- `nextvibe_scan`: scan project files and persist the scan result
- `nextvibe_suggest`: suggest the next bounded task
- `nextvibe_task`: create or read the current task
- `nextvibe_check`: verify completion evidence for the active task

Each tool returns JSON text content and structured content with the same stable
shape used by the CLI JSON commands.

## Smoke Test

You can manually inspect the server with a single request:

```bash
printf '{"jsonrpc":"2.0","id":1,"method":"tools/list"}\n' | nextvibe mcp
```

The response should include `nextvibe_scan`, `nextvibe_suggest`,
`nextvibe_task`, and `nextvibe_check`.
