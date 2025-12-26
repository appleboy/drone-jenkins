# drone-jenkins

[English](README.md) | [繁體中文](README.zh-TW.md) | [简体中文](README.zh-CN.md)

![logo](./images/logo.png)

[![Lint and Testing](https://github.com/appleboy/drone-jenkins/actions/workflows/lint.yml/badge.svg)](https://github.com/appleboy/drone-jenkins/actions/workflows/lint.yml)
[![Trivy Security Scan](https://github.com/appleboy/drone-jenkins/actions/workflows/trivy.yml/badge.svg)](https://github.com/appleboy/drone-jenkins/actions/workflows/trivy.yml)
[![GoDoc](https://godoc.org/github.com/appleboy/drone-jenkins?status.svg)](https://godoc.org/github.com/appleboy/drone-jenkins)
[![codecov](https://codecov.io/gh/appleboy/drone-jenkins/branch/master/graph/badge.svg)](https://codecov.io/gh/appleboy/drone-jenkins)
[![Go Report Card](https://goreportcard.com/badge/github.com/appleboy/drone-jenkins)](https://goreportcard.com/report/github.com/appleboy/drone-jenkins)

一个用于触发 [Jenkins](https://jenkins.io/) 任务的 CLI 工具与 CI/CD 插件。支持 [GitHub Actions](https://github.com/features/actions)、[GitLab CI](https://docs.gitlab.com/ee/ci/)、[Gitea Action](https://docs.gitea.com/usage/actions/overview) 以及任何支持 Docker 容器或 Shell 命令的平台。

## 为什么选择 drone-jenkins？

在现代企业环境中，团队经常根据特定需求、项目要求或历史决策采用不同的 CI/CD 平台。常见的情况包括：

- **多个 CI 平台并存**：有些团队因为 Jenkins 丰富的插件生态系统而使用它，而其他团队则偏好 GitHub Actions 或 GitLab CI 的简洁性和容器原生方式。
- **遗留系统集成**：拥有既有 Jenkins 流水线的组织需要与新的 CI/CD 工作流程集成，而不需要重写所有内容。
- **跨团队协作**：不同部门可能标准化使用不同的工具，需要平台之间的无缝沟通。

**drone-jenkins** 弥补了这个差距，让 CI/CD 流水线能够将触发 Jenkins 任务作为工作流程的一部分。它可以与 **GitHub Actions**、**GitLab CI**、**Gitea Action** 以及任何支持 Docker 容器或 Shell 命令的 CI 平台无缝协作。

这使得以下场景成为可能：

- **统一的部署流水线**：从任何 CI 平台触发现有的 Jenkins 部署任务，无需迁移
- **渐进式迁移**：团队可以逐步迁移到现代 CI 平台，同时继续使用 Jenkins 任务
- **两全其美**：使用 GitHub Actions 或 GitLab CI 进行现代容器化构建，并使用 Jenkins 处理需要特定插件的专门任务
- **集中式编排**：从单一流水线协调跨多个 CI 系统的构建
- **灵活使用**：提供 CLI 可执行文件或 Docker 镜像——根据您的工作流程选择使用方式

无论您是在管理混合 CI/CD 环境还是编排复杂的多平台部署，drone-jenkins 都能提供您所需的连接能力。

## 目录

- [drone-jenkins](#drone-jenkins)
  - [为什么选择 drone-jenkins？](#为什么选择-drone-jenkins)
  - [目录](#目录)
  - [功能特性](#功能特性)
  - [前置条件](#前置条件)
  - [安装](#安装)
    - [下载可执行文件](#下载可执行文件)
    - [从源码构建](#从源码构建)
    - [Docker 镜像](#docker-镜像)
  - [配置](#配置)
    - [Jenkins 服务器设置](#jenkins-服务器设置)
    - [认证](#认证)
    - [参数参考](#参数参考)
  - [使用方式](#使用方式)
    - [命令行](#命令行)
    - [Docker](#docker)
  - [开发](#开发)
    - [构建](#构建)
    - [测试](#测试)
  - [许可证](#许可证)
  - [贡献](#贡献)

## 功能特性

- 触发单个或多个 Jenkins 任务
- 支持 Jenkins 构建参数
- 多种认证方式（API 令牌或远程触发令牌）
- 等待任务完成，可配置轮询间隔和超时时间
- 调试模式，显示详细参数信息并安全遮蔽令牌
- SSL/TLS 支持，可使用自定义 CA 证书（PEM 内容、文件路径或 URL）
- 跨平台支持（Linux、macOS、Windows）
- 提供 CLI 可执行文件或 Docker 镜像

## 前置条件

- Jenkins 服务器（建议版本 2.0 或更新版本）
- 用于认证的 Jenkins API 令牌或远程触发令牌
- 对于 Jenkins 设置，建议使用 Docker，但非必需

## 安装

### 下载可执行文件

预编译的可执行文件可从[发布页面](https://github.com/appleboy/drone-jenkins/releases)下载，支持：

- **Linux**: amd64, 386
- **macOS (Darwin)**: amd64, 386
- **Windows**: amd64, 386

如果已安装 Go，也可以直接安装：

```sh
go install github.com/appleboy/drone-jenkins@latest
```

### 从源码构建

克隆仓库并构建：

```sh
git clone https://github.com/appleboy/drone-jenkins.git
cd drone-jenkins
make build
```

### Docker 镜像

构建 Docker 镜像：

```sh
make docker
```

或拉取预构建的镜像：

```sh
docker pull ghcr.io/appleboy/drone-jenkins
```

## 配置

### Jenkins 服务器设置

使用 Docker 设置 Jenkins 服务器：

```sh
docker run -d -v jenkins_home:/var/jenkins_home -p 8080:8080 -p 50000:50000 --restart=on-failure jenkins/jenkins:slim
```

### 认证

建议使用 Jenkins API 令牌进行认证。创建 API 令牌的步骤：

1. 登录 Jenkins
2. 点击右上角的用户名
3. 选择"安全"
4. 在"API 令牌"下，点击"添加新令牌"
5. 输入名称并点击"生成"
6. 复制生成的令牌

![personal token](./images/personal-token.png)

或者，您可以使用在 Jenkins 任务设置中配置的远程触发令牌。

### 参数参考

| 参数          | CLI 标志             | 环境变量                                        | 必需          | 说明                                                                      |
| ------------- | -------------------- | ----------------------------------------------- | ------------- | ------------------------------------------------------------------------- |
| Host          | `--host`             | `PLUGIN_URL`, `JENKINS_URL`                     | 是            | Jenkins 基础 URL（例如 `http://jenkins.example.com/`）                    |
| User          | `--user`, `-u`       | `PLUGIN_USER`, `JENKINS_USER`                   | 条件式\*      | Jenkins 用户名                                                            |
| Token         | `--token`, `-t`      | `PLUGIN_TOKEN`, `JENKINS_TOKEN`                 | 条件式\*      | Jenkins API 令牌                                                          |
| Remote Token  | `--remote-token`     | `PLUGIN_REMOTE_TOKEN`, `JENKINS_REMOTE_TOKEN`   | 条件式\*      | Jenkins 远程触发令牌                                                      |
| Job           | `--job`, `-j`        | `PLUGIN_JOB`, `JENKINS_JOB`                     | 是            | Jenkins 任务名称 - 可指定多个                                             |
| Parameters    | `--parameters`, `-p` | `PLUGIN_PARAMETERS`, `JENKINS_PARAMETERS`       | 否            | 构建参数，多行 `key=value` 格式（每行一个）                               |
| Insecure      | `--insecure`         | `PLUGIN_INSECURE`, `JENKINS_INSECURE`           | 否            | 允许不安全的 SSL 连接（默认：false）                                      |
| CA Cert       | `--ca-cert`          | `PLUGIN_CA_CERT`, `JENKINS_CA_CERT`             | 否            | 自定义 CA 证书（PEM 内容、文件路径或 HTTP URL）                           |
| Wait          | `--wait`             | `PLUGIN_WAIT`, `JENKINS_WAIT`                   | 否            | 等待任务完成（默认：false）                                               |
| Poll Interval | `--poll-interval`    | `PLUGIN_POLL_INTERVAL`, `JENKINS_POLL_INTERVAL` | 否            | 状态检查间隔（默认：10s）                                                 |
| Timeout       | `--timeout`          | `PLUGIN_TIMEOUT`, `JENKINS_TIMEOUT`             | 否            | 等待任务完成的最长时间（默认：30m）                                       |
| Debug         | `--debug`            | `PLUGIN_DEBUG`, `JENKINS_DEBUG`                 | 否            | 启用调试模式以显示详细参数信息（默认：false）                             |

**认证要求**：您必须提供以下其中一种：

- `user` + `token`（API 令牌认证），或
- `remote-token`（远程触发令牌认证）

**参数格式**：`parameters` 字段接受多行字符串，每行包含一个 `key=value` 配对：

- 每个参数应该在单独一行
- 格式：`KEY=VALUE`（每行一个）
- 空行会自动忽略
- 只有空白的行会被跳过
- 键名会去除前后空白
- 值会保留有意义的空格
- 值可以包含 `=` 符号（第一个 `=` 之后的所有内容都视为值）

## 使用方式

### 命令行

**单个任务：**

```bash
drone-jenkins \
  --host http://jenkins.example.com/ \
  --user appleboy \
  --token XXXXXXXX \
  --job drone-jenkins-plugin
```

**多个任务：**

```bash
drone-jenkins \
  --host http://jenkins.example.com/ \
  --user appleboy \
  --token XXXXXXXX \
  --job drone-jenkins-plugin-1 \
  --job drone-jenkins-plugin-2
```

**带构建参数：**

```bash
drone-jenkins \
  --host http://jenkins.example.com/ \
  --user appleboy \
  --token XXXXXXXX \
  --job my-jenkins-job \
  --parameters $'ENVIRONMENT=production\nVERSION=1.0.0'
```

或使用环境变量：

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

**使用远程令牌认证：**

```bash
drone-jenkins \
  --host http://jenkins.example.com/ \
  --remote-token REMOTE_TOKEN_HERE \
  --job my-jenkins-job
```

**等待任务完成：**

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

**使用调试模式：**

```bash
drone-jenkins \
  --host http://jenkins.example.com/ \
  --user appleboy \
  --token XXXXXXXX \
  --job my-jenkins-job \
  --debug
```

**使用自定义 CA 证书：**

```bash
# 使用文件路径
drone-jenkins \
  --host https://jenkins.example.com/ \
  --user appleboy \
  --token XXXXXXXX \
  --job my-jenkins-job \
  --ca-cert /path/to/ca.pem

# 使用 URL
drone-jenkins \
  --host https://jenkins.example.com/ \
  --user appleboy \
  --token XXXXXXXX \
  --job my-jenkins-job \
  --ca-cert https://example.com/ca-bundle.crt
```

### Docker

**单个任务：**

```bash
docker run --rm \
  -e JENKINS_URL=http://jenkins.example.com/ \
  -e JENKINS_USER=appleboy \
  -e JENKINS_TOKEN=xxxxxxx \
  -e JENKINS_JOB=drone-jenkins-plugin \
  ghcr.io/appleboy/drone-jenkins
```

**多个任务：**

```bash
docker run --rm \
  -e JENKINS_URL=http://jenkins.example.com/ \
  -e JENKINS_USER=appleboy \
  -e JENKINS_TOKEN=xxxxxxx \
  -e JENKINS_JOB=drone-jenkins-plugin-1,drone-jenkins-plugin-2 \
  ghcr.io/appleboy/drone-jenkins
```

**带构建参数：**

```bash
docker run --rm \
  -e JENKINS_URL=http://jenkins.example.com/ \
  -e JENKINS_USER=appleboy \
  -e JENKINS_TOKEN=xxxxxxx \
  -e JENKINS_JOB=my-jenkins-job \
  -e JENKINS_PARAMETERS=$'ENVIRONMENT=production\nVERSION=1.0.0\nBRANCH=main' \
  ghcr.io/appleboy/drone-jenkins
```

**等待任务完成：**

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

**使用调试模式：**

```bash
docker run --rm \
  -e JENKINS_URL=http://jenkins.example.com/ \
  -e JENKINS_USER=appleboy \
  -e JENKINS_TOKEN=xxxxxxx \
  -e JENKINS_JOB=my-jenkins-job \
  -e JENKINS_DEBUG=true \
  ghcr.io/appleboy/drone-jenkins
```

**使用自定义 CA 证书：**

```bash
# 使用挂载的证书文件
docker run --rm \
  -v /path/to/ca.pem:/ca.pem:ro \
  -e JENKINS_URL=https://jenkins.example.com/ \
  -e JENKINS_USER=appleboy \
  -e JENKINS_TOKEN=xxxxxxx \
  -e JENKINS_JOB=my-jenkins-job \
  -e JENKINS_CA_CERT=/ca.pem \
  ghcr.io/appleboy/drone-jenkins

# 使用 URL
docker run --rm \
  -e JENKINS_URL=https://jenkins.example.com/ \
  -e JENKINS_USER=appleboy \
  -e JENKINS_TOKEN=xxxxxxx \
  -e JENKINS_JOB=my-jenkins-job \
  -e JENKINS_CA_CERT=https://example.com/ca-bundle.crt \
  ghcr.io/appleboy/drone-jenkins
```

更多详细示例和高级配置，请参阅 [DOCS.md](DOCS.md)。

## 开发

### 构建

构建可执行文件：

```sh
make build
```

构建 Docker 镜像：

```sh
make docker
```

### 测试

运行测试套件：

```sh
make test
```

运行测试并生成覆盖率报告：

```sh
make test-coverage
```

## 许可证

Copyright (c) 2019 Bo-Yi Wu

## 贡献

欢迎贡献！请随时提交 Pull Request。
