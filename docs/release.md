# Release

NextVibe is distributed as a small Go CLI. The release path is designed for
open-source use across Windows, macOS, and Linux without requiring hosted
infrastructure beyond GitHub Actions.

## Continuous Integration

The CI workflow runs on:

- Windows
- macOS
- Linux

It also cross-compiles the CLI for:

- `linux/amd64`
- `linux/arm64`
- `darwin/amd64`
- `darwin/arm64`
- `windows/amd64`
- `windows/arm64`

This keeps normal tests tied to real operating systems while still checking the
release build targets on every push and pull request.

## Tagged Releases

To create a release, push a version tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The release workflow builds and attaches these archives:

```text
nextvibe_linux_amd64.tar.gz
nextvibe_linux_arm64.tar.gz
nextvibe_darwin_amd64.tar.gz
nextvibe_darwin_arm64.tar.gz
nextvibe_windows_amd64.zip
nextvibe_windows_arm64.zip
```

The workflow uses GitHub's default `GITHUB_TOKEN`. No extra repository secrets
are required.

## npm Publishing

The npm package is published as:

```text
@crescentr/nextvibe
```

npm installs do not compile Go. During `postinstall`, the package downloads the
matching prebuilt binary from the GitHub Release whose tag matches the package
version. For example, `@crescentr/nextvibe@0.1.0` downloads from `v0.1.0`.

Publish in this order:

```bash
git tag v0.1.0
git push origin v0.1.0
# wait for the GitHub Release assets to finish uploading
npm publish --access public
```

Use a dry run before publishing:

```bash
npm publish --dry-run --access public
```

The release asset names are part of the npm install contract, so keep these
names stable unless `npm/postinstall.js` is updated in the same release:

```text
nextvibe_linux_amd64.tar.gz
nextvibe_linux_arm64.tar.gz
nextvibe_darwin_amd64.tar.gz
nextvibe_darwin_arm64.tar.gz
nextvibe_windows_amd64.zip
nextvibe_windows_arm64.zip
```

## Manual Build Check

Maintainers can verify the same build path locally with Go 1.22 or newer:

```bash
go test ./...
mkdir -p dist
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/nextvibe-linux-amd64 ./cmd/nextvibe
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o dist/nextvibe-darwin-arm64 ./cmd/nextvibe
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/nextvibe-windows-amd64.exe ./cmd/nextvibe
```
