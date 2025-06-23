# humpback-agent

[![PkgGoDev](https://pkg.go.dev/badge/github.com/docker/docker)](https://golang.org/)
[![Docker](https://img.shields.io/badge/docker-pull-blue?logo=docker)](https://hub.docker.com/r/humpbacks/humpback-agent)
[![Base: Moby](https://img.shields.io/badge/Base-Moby-2496ED?logo=docker&logoColor=white)](https://github.com/moby/moby)
[![Release](https://img.shields.io/badge/release-v2.0.0-blue)](https://github.com/humpback/humpback-agent/releases/tag/v2.0.0)

![Humpback logo](/assets/logo.png)

The service executor of Humpback, which provides container operations and cron execution for Humpback.

## language

- [English](README.md)
- [中文](README.zh.md)

## Feature

- Heartbeat debriefing.
- Container operations。
- Support for cron.

## Getting Started

* [Humpback Guides](https://humpback.github.io/humpback)

## Installing

By default, Humpback Agent will expose a API server over port `8018` for receiving Humpback Server call.

```bash

docker run -d \
  --name=humpback-agent \
  --net=host \
  --restart=always \
  --privileged \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /var/lib/docker:/var/lib/docker \
  -e HUMPBACK_SERVER_REGISTER_TOKEN={token} \
  -e HUMPBACK_SERVER_HOST={server-address}:8101 \
  -e HUMPBACK_VOLUMES_ROOT_DIRECTORY=/var/lib/docker \
  humpbacks/humpback-agent

```

Please replace `{server-address}` to the Humbpack Server IP.

## Usage

After the installation is completed, add the current machine IP address to the **Nodes** page, and you can schedule it after the status changes to **Healthy**.

![Nodes](/assets/nodes.png)

## Licence

Humpback Server is licensed under the [Apache Licence 2.0](http://www.apache.org/licenses/LICENSE-2.0.html).   
