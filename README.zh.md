# humpback-agent

[![PkgGoDev](https://pkg.go.dev/badge/github.com/docker/docker)](https://golang.org/)
[![Docker](https://img.shields.io/badge/docker-pull-blue?logo=docker)](https://hub.docker.com/r/humpbacks/humpback-agent)
[![Base: Moby](https://img.shields.io/badge/Base-Moby-2496ED?logo=docker&logoColor=white)](https://github.com/moby/moby)
[![Release](https://img.shields.io/badge/release-v2.0.0-blue)](https://github.com/humpback/humpback-agent/releases/tag/v2.0.0)

![Humpback logo](/assets/logo.png)

Humpback的服务执行程序，为Humpback提供容器操作和Cron执行。

## 语言

- [English](README.md)
- [中文](README.zh.md)

## 特征

- 心跳汇报。
- 容器操作。
- 支持Cron.

## 快速开始

* [Humpback Guides](https://humpback.github.io/humpback)

## 安装

Humpback Agent默认会监听8018端口用于接收Humpback Server的调用。

```bash

docker run -d \
  --name=humpback-agent \
  --net=host \
  --restart=always \
  --privileged \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /var/lib/docker:/var/lib/docker \
  -e HUMPBACK_AGENT_API_BIND=0.0.0.0:8018 \
  -e HUMPBACK_SERVER_HOST={server-address}:8101 \
  -e HUMPBACK_VOLUMES_ROOT_DIRECTORY=/var/lib/docker \
  humpbacks/humpback-agent

```

请注意：将{server-address}替换为部署Humpback Server的真实IP地址。

## 使用

安装完成后，将当前机器IP地址添加到**机器管理**页面，待状态变为**在线**后即可进行调度使用。

![Nodes](/assets/nodes-zh.png)

## 许可证

Humpback 根据 [Apache Licence 2.0](http://www.apache.org/licenses/LICENSE-2.0.html) 获得许可。
