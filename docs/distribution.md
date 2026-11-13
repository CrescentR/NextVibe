# Distribution

NextVibe now includes package-manager surfaces for npm, Homebrew, and Scoop.
They are intentionally local-first and do not require publishing credentials to
run verification.

## npm

The npm package entrypoint is `npm/nextvibe.js`.

The npm package installs a prebuilt binary during `postinstall`. It maps the
package version to a GitHub Release tag, so `@crescentr/nextvibe@0.1.0`
downloads assets from `v0.1.0`.

Local install test:

```bash
npm install -g .
nextvibe --help
```

For local package tests before a GitHub Release exists, build a binary and point
the installer at it:

```bash
go build -o dist/nextvibe ./cmd/nextvibe
NEXTVIBE_INSTALL_BINARY="$PWD/dist/nextvibe" npm install -g .
```

Publish order matters:

1. Push the version tag and let GitHub Actions upload all release archives.
2. Run `npm publish --access public`.

Do not publish npm before the matching GitHub Release assets exist, because npm
users download those assets during install.

## Homebrew

The formula lives at:

```text
Formula/nextvibe.rb
```

Before publishing to a tap, replace:

```text
REPLACE_WITH_RELEASE_TARBALL_SHA256
```

with the SHA-256 of the tagged source tarball.

## Scoop

The Scoop manifest lives at:

```text
scoop/nextvibe.json
```

Before publishing to a bucket, replace:

```text
REPLACE_WITH_WINDOWS_AMD64_SHA256
REPLACE_WITH_WINDOWS_ARM64_SHA256
```

with the checksums from `dist/checksums.txt` for the matching release archives.

## Release Archives

The release workflow produces:

```text
nextvibe_linux_amd64.tar.gz
nextvibe_linux_arm64.tar.gz
nextvibe_darwin_amd64.tar.gz
nextvibe_darwin_arm64.tar.gz
nextvibe_windows_amd64.zip
nextvibe_windows_arm64.zip
checksums.txt
```
