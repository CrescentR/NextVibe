# Deployment

NextVibe can be packaged as a local Docker image that runs the CLI directly.
This path is intentionally local-only and does not require a cloud account,
hosted registry, or secrets.

## Build The Image

From the repository root:

```bash
docker build -t nextvibe:local .
```

The Dockerfile uses a Go builder image to compile `./cmd/nextvibe`, then copies
the compiled binary into a minimal runtime image.

## Run Against A Project

Mount the project you want to inspect at `/workspace` and pass normal NextVibe
CLI arguments after the image name.

macOS or Linux:

```bash
docker run --rm -v "$PWD:/workspace" -w /workspace nextvibe:local scan --json
docker run --rm -v "$PWD:/workspace" -w /workspace nextvibe:local suggest --json
docker run --rm -v "$PWD:/workspace" -w /workspace nextvibe:local task --json
docker run --rm -v "$PWD:/workspace" -w /workspace nextvibe:local check --json
```

PowerShell:

```powershell
docker run --rm -v "${PWD}:/workspace" -w /workspace nextvibe:local scan --json
docker run --rm -v "${PWD}:/workspace" -w /workspace nextvibe:local suggest --json
docker run --rm -v "${PWD}:/workspace" -w /workspace nextvibe:local task --json
docker run --rm -v "${PWD}:/workspace" -w /workspace nextvibe:local check --json
```

## Local Verification

After building the image, verify the packaged CLI:

```bash
docker run --rm nextvibe:local --help
```

Then run the mounted-project commands above from a project root.
