# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

drone-jenkins is a Drone CI plugin (and standalone CLI tool) for triggering Jenkins jobs. It supports multiple authentication methods, build parameters, and can wait for job completion with configurable polling.

## Build Commands

```sh
make build       # Build binary to bin/drone-jenkins
make test        # Run tests with coverage
make lint        # Run golangci-lint
make fmt         # Format code with golangci-lint
make docker      # Build Docker image
make clean       # Clean build artifacts
```

To run a single test:

```sh
go test -v -run TestFunctionName ./...
```

## Architecture

The codebase is structured as a simple Go CLI application:

- **main.go** - CLI entry point using `urfave/cli/v2`. Defines all command-line flags and environment variable mappings. Handles debug mode display with token masking.

- **plugin.go** - Plugin struct and configuration validation. Contains `Exec()` which orchestrates job triggering. Includes `parseParameters()` for converting multi-line `key=value` strings to URL values.

- **jenkins.go** - Jenkins HTTP client implementation. Handles:
  - Authentication (basic auth with API token)
  - SSL/TLS with custom CA certificates (PEM content, file path, or URL)
  - Job triggering via `/build` or `/buildWithParameters` endpoints
  - Queue monitoring and build status polling for wait mode
  - Nested job path parsing (converts `folder/job` to `/job/folder/job/job`)

## Key Patterns

- **Authentication**: Either `user + token` (API token auth) OR `remote-token` (remote trigger token). Validated in `main.go:run()`.

- **Parameters format**: Multi-line string with one `key=value` per line. Parsed in `plugin.go:parseParameters()`.

- **Wait mode**: Uses two-phase polling - first waits for queue item to get a build number, then polls build status until completion.

- **Environment variables**: Each flag accepts multiple env vars (e.g., `PLUGIN_URL`, `JENKINS_URL`, `INPUT_URL`) for compatibility with different CI systems.
