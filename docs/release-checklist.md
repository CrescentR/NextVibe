# Release Checklist

Use this checklist before cutting a NextVibe release.

## Local Quality Gate

Run:

```bash
go test ./...
go build ./...
git diff --check
sh scripts/verify-release.sh
```

On Windows:

```powershell
go test ./...
go build ./...
git diff --check
./scripts/verify-release.ps1
```

## Protocol Gate

Verify the agent-facing surfaces:

```bash
nextvibe scan --json
nextvibe suggest --json
nextvibe task --json
nextvibe check --json
nextvibe install verify --json
```

For MCP:

```bash
printf '{"jsonrpc":"2.0","id":1,"method":"tools/list"}\n' | nextvibe mcp
```

## Distribution Gate

Before publishing package-manager entries:

- replace the Homebrew formula SHA-256 placeholder
- replace the Scoop manifest hash placeholders
- verify `npm install -g .` on a machine with Go 1.22 or newer
- confirm the GitHub release contains all six archives and `checksums.txt`
