# drone-jenkins

[English](README.md) | [繁體中文](README.zh-TW.md) | [简体中文](README.zh-CN.md)

![logo](./images/logo.png)

[![Lint and Testing](https://github.com/appleboy/drone-jenkins/actions/workflows/lint.yml/badge.svg)](https://github.com/appleboy/drone-jenkins/actions/workflows/lint.yml)
[![Trivy Security Scan](https://github.com/appleboy/drone-jenkins/actions/workflows/trivy.yml/badge.svg)](https://github.com/appleboy/drone-jenkins/actions/workflows/trivy.yml)
[![GoDoc](https://godoc.org/github.com/appleboy/drone-jenkins?status.svg)](https://godoc.org/github.com/appleboy/drone-jenkins)
[![codecov](https://codecov.io/gh/appleboy/drone-jenkins/branch/master/graph/badge.svg)](https://codecov.io/gh/appleboy/drone-jenkins)
[![Go Report Card](https://goreportcard.com/badge/github.com/appleboy/drone-jenkins)](https://goreportcard.com/report/github.com/appleboy/drone-jenkins)

A CLI tool and CI/CD plugin for triggering [Jenkins](https://jenkins.io/) jobs. Works with [GitHub Actions](https://github.com/features/actions), [GitLab CI](https://docs.gitlab.com/ee/ci/), [Gitea Action](https://docs.gitea.com/usage/actions/overview), and any platform that supports Docker containers or shell commands.

## Why drone-jenkins?

In modern enterprise environments, teams often adopt different CI/CD platforms based on their specific needs, project requirements, or historical decisions. It's common to find:

- **Multiple CI platforms coexisting**: Some teams use Jenkins for its extensive plugin ecosystem, while others prefer GitHub Actions or GitLab CI for their simplicity and container-native approach.
- **Legacy systems integration**: Organizations with established Jenkins pipelines need to integrate with newer CI/CD workflows without rewriting everything.
- **Cross-team collaboration**: Different departments may standardize on different tools, requiring seamless communication between platforms.

**drone-jenkins** bridges this gap by allowing CI/CD pipelines to trigger Jenkins jobs as part of their workflow. It works seamlessly with **GitHub Actions**, **GitLab CI**, **Gitea Action**, and any CI platform that supports Docker containers or shell commands.

This enables:

- **Unified deployment pipelines**: Trigger existing Jenkins deployment jobs from any CI platform without migration
- **Gradual migration**: Teams can incrementally move to modern CI platforms while still leveraging Jenkins jobs
- **Best of both worlds**: Use GitHub Actions or GitLab CI for modern containerized builds and Jenkins for specialized tasks with specific plugins
- **Centralized orchestration**: Coordinate builds across multiple CI systems from a single pipeline
- **Flexibility**: Available as a CLI binary or Docker image—use it however fits your workflow

Whether you're managing a hybrid CI/CD environment or orchestrating complex multi-platform deployments, drone-jenkins provides the connectivity you need.

## Table of Contents

- [drone-jenkins](#drone-jenkins)
  - [Why drone-jenkins?](#why-drone-jenkins)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
    - [Download Binary](#download-binary)
    - [Build from Source](#build-from-source)
    - [Docker Image](#docker-image)
  - [Configuration](#configuration)
    - [Jenkins Server Setup](#jenkins-server-setup)
    - [Authentication](#authentication)
    - [Parameters Reference](#parameters-reference)
  - [Usage](#usage)
    - [Command Line](#command-line)
    - [Docker](#docker)
  - [Development](#development)
    - [Building](#building)
    - [Testing](#testing)
  - [License](#license)
  - [Contributing](#contributing)

## Features

- Trigger single or multiple Jenkins jobs
- Support for Jenkins build parameters
- Multiple authentication methods (API token or remote trigger token)
- Wait for job completion with configurable polling and timeout
- Debug mode with detailed parameter information and secure token masking
- SSL/TLS support with custom CA certificates (PEM content, file path, or URL)
- Cross-platform support (Linux, macOS, Windows)
- Available as CLI binary or Docker image

## Prerequisites

- Jenkins server (version 2.0 or later recommended)
- Jenkins API token or remote trigger token for authentication
- For Jenkins setup, Docker is recommended but not required

## Installation

### Download Binary

Pre-compiled binaries are available from the [release page](https://github.com/appleboy/drone-jenkins/releases) for:

- **Linux**: amd64, 386
- **macOS (Darwin)**: amd64, 386
- **Windows**: amd64, 386

With Go installed, you can also install directly:

```sh
go install github.com/appleboy/drone-jenkins@latest
```

### Build from Source

Clone the repository and build:

```sh
git clone https://github.com/appleboy/drone-jenkins.git
cd drone-jenkins
make build
```

### Docker Image

Build the Docker image:

```sh
make docker
```

Or pull the pre-built image:

```sh
docker pull ghcr.io/appleboy/drone-jenkins
```

## Configuration

### Jenkins Server Setup

Set up a Jenkins server using Docker:

```sh
docker run -d -v jenkins_home:/var/jenkins_home -p 8080:8080 -p 50000:50000 --restart=on-failure jenkins/jenkins:slim
```

### Authentication

Jenkins API tokens are recommended for authentication. To create an API token:

1. Log into Jenkins
2. Click on your username (top right)
3. Select "Security"
4. Under "API Token", click "Add new Token"
5. Give it a name and click "Generate"
6. Copy the generated token

![personal token](./images/personal-token.png)

Alternatively, you can use a remote trigger token configured in your Jenkins job settings.

### Parameters Reference

| Parameter     | CLI Flag             | Environment Variable                            | Required      | Description                                                               |
| ------------- | -------------------- | ----------------------------------------------- | ------------- | ------------------------------------------------------------------------- |
| Host          | `--host`             | `PLUGIN_URL`, `JENKINS_URL`                     | Yes           | Jenkins base URL (e.g., `http://jenkins.example.com/`)                    |
| User          | `--user`, `-u`       | `PLUGIN_USER`, `JENKINS_USER`                   | Conditional\* | Jenkins username                                                          |
| Token         | `--token`, `-t`      | `PLUGIN_TOKEN`, `JENKINS_TOKEN`                 | Conditional\* | Jenkins API token                                                         |
| Remote Token  | `--remote-token`     | `PLUGIN_REMOTE_TOKEN`, `JENKINS_REMOTE_TOKEN`   | Conditional\* | Jenkins remote trigger token                                              |
| Job           | `--job`, `-j`        | `PLUGIN_JOB`, `JENKINS_JOB`                     | Yes           | Jenkins job name(s) - can specify multiple                                |
| Parameters    | `--parameters`, `-p` | `PLUGIN_PARAMETERS`, `JENKINS_PARAMETERS`       | No            | Build parameters in multi-line `key=value` format (one per line)          |
| Insecure      | `--insecure`         | `PLUGIN_INSECURE`, `JENKINS_INSECURE`           | No            | Allow insecure SSL connections (default: false)                           |
| CA Cert       | `--ca-cert`          | `PLUGIN_CA_CERT`, `JENKINS_CA_CERT`             | No            | Custom CA certificate (PEM content, file path, or HTTP URL)               |
| Wait          | `--wait`             | `PLUGIN_WAIT`, `JENKINS_WAIT`                   | No            | Wait for job completion (default: false)                                  |
| Poll Interval | `--poll-interval`    | `PLUGIN_POLL_INTERVAL`, `JENKINS_POLL_INTERVAL` | No            | Interval between status checks (default: 10s)                             |
| Timeout       | `--timeout`          | `PLUGIN_TIMEOUT`, `JENKINS_TIMEOUT`             | No            | Maximum time to wait for job completion (default: 30m)                    |
| Debug         | `--debug`            | `PLUGIN_DEBUG`, `JENKINS_DEBUG`                 | No            | Enable debug mode to show detailed parameter information (default: false) |

**Authentication Requirements**: You must provide either:

- `user` + `token` (API token authentication), OR
- `remote-token` (remote trigger token authentication)

**Parameters Format**: The `parameters` field accepts a multi-line string where each line contains one `key=value` pair:

- Each parameter should be on a separate line
- Format: `KEY=VALUE` (one per line)
- Empty lines are automatically ignored
- Whitespace-only lines are skipped
- Keys are trimmed of surrounding whitespace
- Values preserve intentional spaces
- Values can contain `=` signs (everything after the first `=` is treated as the value)

## Usage

### Command Line

**Single job:**

```bash
drone-jenkins \
  --host http://jenkins.example.com/ \
  --user appleboy \
  --token XXXXXXXX \
  --job drone-jenkins-plugin
```

**Multiple jobs:**

```bash
drone-jenkins \
  --host http://jenkins.example.com/ \
  --user appleboy \
  --token XXXXXXXX \
  --job drone-jenkins-plugin-1 \
  --job drone-jenkins-plugin-2
```

**With build parameters:**

```bash
drone-jenkins \
  --host http://jenkins.example.com/ \
  --user appleboy \
  --token XXXXXXXX \
  --job my-jenkins-job \
  --parameters $'ENVIRONMENT=production\nVERSION=1.0.0'
```

Or using environment variable:

```bash
export JENKINS_PARAMETERS="ENVIRONMENT=production
VERSION=1.0.0
BRANCH=main"

drone-jenkins \
  --host http://jenkins.example.com/ \
  --user appleboy \
  --token XXXXXXXX \
  --job my-jenkins-job
```

**Using remote token authentication:**

```bash
drone-jenkins \
  --host http://jenkins.example.com/ \
  --remote-token REMOTE_TOKEN_HERE \
  --job my-jenkins-job
```

**Wait for job completion:**

```bash
drone-jenkins \
  --host http://jenkins.example.com/ \
  --user appleboy \
  --token XXXXXXXX \
  --job my-jenkins-job \
  --wait \
  --poll-interval 15s \
  --timeout 1h
```

**With debug mode:**

```bash
drone-jenkins \
  --host http://jenkins.example.com/ \
  --user appleboy \
  --token XXXXXXXX \
  --job my-jenkins-job \
  --debug
```

**With custom CA certificate:**

```bash
# Using a file path
drone-jenkins \
  --host https://jenkins.example.com/ \
  --user appleboy \
  --token XXXXXXXX \
  --job my-jenkins-job \
  --ca-cert /path/to/ca.pem

# Using a URL
drone-jenkins \
  --host https://jenkins.example.com/ \
  --user appleboy \
  --token XXXXXXXX \
  --job my-jenkins-job \
  --ca-cert https://example.com/ca-bundle.crt
```

### Docker

**Single job:**

```bash
docker run --rm \
  -e JENKINS_URL=http://jenkins.example.com/ \
  -e JENKINS_USER=appleboy \
  -e JENKINS_TOKEN=xxxxxxx \
  -e JENKINS_JOB=drone-jenkins-plugin \
  ghcr.io/appleboy/drone-jenkins
```

**Multiple jobs:**

```bash
docker run --rm \
  -e JENKINS_URL=http://jenkins.example.com/ \
  -e JENKINS_USER=appleboy \
  -e JENKINS_TOKEN=xxxxxxx \
  -e JENKINS_JOB=drone-jenkins-plugin-1,drone-jenkins-plugin-2 \
  ghcr.io/appleboy/drone-jenkins
```

**With build parameters:**

```bash
docker run --rm \
  -e JENKINS_URL=http://jenkins.example.com/ \
  -e JENKINS_USER=appleboy \
  -e JENKINS_TOKEN=xxxxxxx \
  -e JENKINS_JOB=my-jenkins-job \
  -e JENKINS_PARAMETERS=$'ENVIRONMENT=production\nVERSION=1.0.0\nBRANCH=main' \
  ghcr.io/appleboy/drone-jenkins
```

**Wait for job completion:**

```bash
docker run --rm \
  -e JENKINS_URL=http://jenkins.example.com/ \
  -e JENKINS_USER=appleboy \
  -e JENKINS_TOKEN=xxxxxxx \
  -e JENKINS_JOB=my-jenkins-job \
  -e JENKINS_WAIT=true \
  -e JENKINS_POLL_INTERVAL=15s \
  -e JENKINS_TIMEOUT=1h \
  ghcr.io/appleboy/drone-jenkins
```

**With debug mode:**

```bash
docker run --rm \
  -e JENKINS_URL=http://jenkins.example.com/ \
  -e JENKINS_USER=appleboy \
  -e JENKINS_TOKEN=xxxxxxx \
  -e JENKINS_JOB=my-jenkins-job \
  -e JENKINS_DEBUG=true \
  ghcr.io/appleboy/drone-jenkins
```

**With custom CA certificate:**

```bash
# Using a mounted certificate file
docker run --rm \
  -v /path/to/ca.pem:/ca.pem:ro \
  -e JENKINS_URL=https://jenkins.example.com/ \
  -e JENKINS_USER=appleboy \
  -e JENKINS_TOKEN=xxxxxxx \
  -e JENKINS_JOB=my-jenkins-job \
  -e JENKINS_CA_CERT=/ca.pem \
  ghcr.io/appleboy/drone-jenkins

# Using a URL
docker run --rm \
  -e JENKINS_URL=https://jenkins.example.com/ \
  -e JENKINS_USER=appleboy \
  -e JENKINS_TOKEN=xxxxxxx \
  -e JENKINS_JOB=my-jenkins-job \
  -e JENKINS_CA_CERT=https://example.com/ca-bundle.crt \
  ghcr.io/appleboy/drone-jenkins
```

For more detailed examples and advanced configurations, see [DOCS.md](DOCS.md).

## Development

### Building

Build the binary:

```sh
make build
```

Build the Docker image:

```sh
make docker
```

### Testing

Run the test suite:

```sh
make test
```

Run tests with coverage:

```sh
make test-coverage
```

## License

Copyright (c) 2019 Bo-Yi Wu

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
