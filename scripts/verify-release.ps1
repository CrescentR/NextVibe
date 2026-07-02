$ErrorActionPreference = "Stop"

go test ./...
go build ./...

$tmp = Join-Path ([System.IO.Path]::GetTempPath()) "nextvibe-release-check.exe"
go build -trimpath -o $tmp ./cmd/nextvibe
& $tmp --help | Out-Null

$required = @(
  "package.json",
  "npm/nextvibe.js",
  "npm/postinstall.js",
  "Formula/nextvibe.rb",
  "scoop/nextvibe.json",
  "docs/mcp.md",
  "docs/distribution.md",
  "docs/release-checklist.md"
)

foreach ($file in $required) {
  if (-not (Test-Path $file)) {
    throw "missing release file: $file"
  }
}

Write-Host "release verification passed"
