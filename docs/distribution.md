# Distribution

NextVibe now includes package-manager surfaces for npm, Homebrew, and Scoop.
They are intentionally local-first and do not require publishing credentials to
run verification.

## npm

The npm package entrypoint is `npm/nextvibe.js`.

Local install test:

```bash
npm install -g .
nextvibe --help
```

The package currently builds the Go binary during `postinstall`, so npm users
need Go 1.22 or newer on `PATH`. A later binary-download postinstall can replace
this without changing the CLI contract.

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
