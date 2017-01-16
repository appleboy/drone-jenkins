# drone-jenkins

[![Build Status](https://travis-ci.org/appleboy/drone-jenkins.svg?branch=master)](https://travis-ci.org/appleboy/drone-jenkins) [![codecov](https://codecov.io/gh/appleboy/drone-jenkins/branch/master/graph/badge.svg)](https://codecov.io/gh/appleboy/drone-jenkins) [![Go Report Card](https://goreportcard.com/badge/github.com/appleboy/drone-jenkins)](https://goreportcard.com/report/github.com/appleboy/drone-jenkins)

[Drone](https://github.com/drone/drone) plugin for trigger [Jenkins](https://jenkins.io/) jobs.

## Build

Build the binary with the following commands:

```
$ make build
```

## Testing

Test the package with the following command:

```
$ make test
```

## Docker

Build the docker image with the following commands:

```
$ make docker
```

Please note incorrectly building the image for the correct x64 linux and with
GCO disabled will result in an error when running the Docker image:

```
docker: Error response from daemon: Container command
'/bin/drone-jenkins' not found or does not exist..
```

## Usage

Execute from the working directory:

```
docker run --rm \
  -e PLUGIN_BASE_URL=http://example.com \
  -e PLUGIN_USERNAME=xxxxxxx \
  -e PLUGIN_TOKEN=xxxxxxx \
  -e PLUGIN_JOB=xxxxxxx \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  appleboy/drone-jenkins
```

Load all environments from file.

```bash
docker run --rm \
  -e ENV_FILE=your_env_file_path \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  appleboy/drone-jenkins
```
