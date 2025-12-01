# drone-jenkins

![logo](./images/logo.png)

[![Lint and Testing](https://github.com/appleboy/drone-jenkins/actions/workflows/lint.yml/badge.svg)](https://github.com/appleboy/drone-jenkins/actions/workflows/lint.yml)
[![Trivy Security Scan](https://github.com/appleboy/drone-jenkins/actions/workflows/trivy.yml/badge.svg)](https://github.com/appleboy/drone-jenkins/actions/workflows/trivy.yml)
[![GoDoc](https://godoc.org/github.com/appleboy/drone-jenkins?status.svg)](https://godoc.org/github.com/appleboy/drone-jenkins)
[![codecov](https://codecov.io/gh/appleboy/drone-jenkins/branch/master/graph/badge.svg)](https://codecov.io/gh/appleboy/drone-jenkins)
[![Go Report Card](https://goreportcard.com/badge/github.com/appleboy/drone-jenkins)](https://goreportcard.com/report/github.com/appleboy/drone-jenkins)

A [Drone](https://github.com/drone/drone) plugin for triggering [Jenkins](https://jenkins.io/) jobs with flexible authentication and parameter support.

## Table of Contents

- [drone-jenkins](#drone-jenkins)
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
    - [Drone CI](#drone-ci)
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
- SSL/TLS support with optional insecure mode
- Cross-platform support (Linux, macOS, Windows)
- Available as binary, Docker image, or Drone plugin

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
docker run \
  --name jenkins \
  -d --restart always \
  -p 8080:8080 -p 50000:50000 \
  -v /data/jenkins:/var/jenkins_home \
  jenkins/jenkins:lts
```

**Note**: Create the `/data/jenkins` directory before starting Jenkins.

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

| Parameter     | CLI Flag             | Environment Variable                            | Required      | Description                                            |
| ------------- | -------------------- | ----------------------------------------------- | ------------- | ------------------------------------------------------ |
| Host          | `--host`             | `PLUGIN_URL`, `JENKINS_URL`                     | Yes           | Jenkins base URL (e.g., `http://jenkins.example.com/`) |
| User          | `--user`, `-u`       | `PLUGIN_USER`, `JENKINS_USER`                   | Conditional\* | Jenkins username                                       |
| Token         | `--token`, `-t`      | `PLUGIN_TOKEN`, `JENKINS_TOKEN`                 | Conditional\* | Jenkins API token                                      |
| Remote Token  | `--remote-token`     | `PLUGIN_REMOTE_TOKEN`, `JENKINS_REMOTE_TOKEN`   | Conditional\* | Jenkins remote trigger token                           |
| Job           | `--job`, `-j`        | `PLUGIN_JOB`, `JENKINS_JOB`                     | Yes           | Jenkins job name(s) - can specify multiple             |
| Parameters    | `--parameters`, `-p` | `PLUGIN_PARAMETERS`, `JENKINS_PARAMETERS`       | No            | Build parameters in `key=value` format                 |
| Insecure      | `--insecure`         | `PLUGIN_INSECURE`, `JENKINS_INSECURE`           | No            | Allow insecure SSL connections (default: false)        |
| Wait          | `--wait`             | `PLUGIN_WAIT`, `JENKINS_WAIT`                   | No            | Wait for job completion (default: false)               |
| Poll Interval | `--poll-interval`    | `PLUGIN_POLL_INTERVAL`, `JENKINS_POLL_INTERVAL` | No            | Interval between status checks (default: 10s)          |
| Timeout       | `--timeout`          | `PLUGIN_TIMEOUT`, `JENKINS_TIMEOUT`             | No            | Maximum time to wait for job completion (default: 30m) |

**Authentication Requirements**: You must provide either:

- `user` + `token` (API token authentication), OR
- `remote-token` (remote trigger token authentication)

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
  --parameters "ENVIRONMENT=production" \
  --parameters "VERSION=1.0.0"
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
  -e JENKINS_PARAMETERS="ENVIRONMENT=production,VERSION=1.0.0" \
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

### Drone CI

Add the plugin to your `.drone.yml`:

```yaml
kind: pipeline
name: default

steps:
  - name: trigger-jenkins
    image: ghcr.io/appleboy/drone-jenkins
    settings:
      url: http://jenkins.example.com/
      user: appleboy
      token:
        from_secret: jenkins_token
      job: drone-jenkins-plugin
```

**Multiple jobs with parameters:**

```yaml
steps:
  - name: trigger-jenkins
    image: ghcr.io/appleboy/drone-jenkins
    settings:
      url: http://jenkins.example.com/
      user: appleboy
      token:
        from_secret: jenkins_token
      job:
        - deploy-frontend
        - deploy-backend
      parameters:
        - ENVIRONMENT=production
        - VERSION=${DRONE_TAG}
        - COMMIT_SHA=${DRONE_COMMIT_SHA}
```

**Using remote token:**

```yaml
steps:
  - name: trigger-jenkins
    image: ghcr.io/appleboy/drone-jenkins
    settings:
      url: http://jenkins.example.com/
      remote_token:
        from_secret: jenkins_remote_token
      job: my-jenkins-job
```

**Wait for job completion:**

```yaml
steps:
  - name: trigger-jenkins
    image: ghcr.io/appleboy/drone-jenkins
    settings:
      url: http://jenkins.example.com/
      user: appleboy
      token:
        from_secret: jenkins_token
      job: deploy-production
      wait: true
      poll_interval: 15s
      timeout: 1h
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
