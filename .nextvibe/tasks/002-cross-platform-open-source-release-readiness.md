# 002 - Add cross-platform open-source release readiness

Task ID: 002
Status: complete
Title: Add cross-platform open-source release readiness

## Background

The project is intended to be open source and usable from Windows, macOS, and Linux.
It already builds as a Go CLI, but it needs repeatable platform checks, release
artifacts, and installation guidance before it is comfortable for public users.

## Goal

Create a minimal open-source release path that verifies the CLI on Windows,
macOS, and Linux and documents how users can install or build it locally.

## Allowed Files

- .github/
- .github/workflows/ci.yml
- .github/workflows/release.yml
- .gitattributes
- .gitignore
- README.md
- docs/development-roadmap.md
- docs/usage.md
- docs/release.md
- docs/roadmap.md
- Dockerfile
- .dockerignore
- docs/deployment.md

## Forbidden Changes

- do not change CLI behavior
- do not add cloud provider resources
- do not introduce secrets
- do not add runtime dependencies
- do not expand into packaging systems such as Homebrew, winget, or apt yet

## Acceptance Criteria

- CI runs tests on Windows, macOS, and Linux
- CI verifies cross-compilation for common Windows, macOS, and Linux targets
- Tagged releases can publish platform-specific CLI archives
- README and usage docs explain platform support and local installation
- Release docs describe a maintainer workflow that does not require secrets beyond GitHub's default token
