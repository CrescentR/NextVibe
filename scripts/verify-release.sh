#!/usr/bin/env sh
set -eu

go test ./...
go build ./...

tmp="${TMPDIR:-/tmp}/nextvibe-release-check"
go build -trimpath -o "$tmp" ./cmd/nextvibe
"$tmp" --help >/dev/null

for file in \
  package.json \
  npm/nextvibe.js \
  npm/postinstall.js \
  Formula/nextvibe.rb \
  scoop/nextvibe.json \
  docs/mcp.md \
  docs/distribution.md \
  docs/release-checklist.md
do
  test -f "$file"
done

echo "release verification passed"
