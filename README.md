## humpback-agent

Humpback Agent 主要为 [Humpback](https://github.com/humpback/humpback) 平台提供容器访问功能。[API 文档](https://github.com/humpback/humpback-agent/wiki)

## How to develop

### Prepare

1. First, you need an server that installed and already running docker(recommend Linux).
2. `[Optional]` If you are develop on windows, you should expose docker host with tcp protocol(default is `unix:///var/run/docker.sock`). How to do it? Run: `docker run -d -v /var/run/docker.sock:/var/run/docker.sock -p 0.0.0.0:1234:1234 bobrik/socat TCP-LISTEN:1234,fork UNIX-CONNECT:/var/run/docker.sock` expose your docker host at `tcp:<ip>:1234`.

### Init dependencies

1. Use `go get` install the dependencies.

**Notice: GOPATH is very important!**

### Build & Run

1. `[Optional]`, If you are develop on windows, change the `conf/app.conf DOCKER_ENDPOINT` config.

```bash
# For example
DOCKER_ENDPOINT = tcp://192.168.1.200:1234
# DOCKER_ENDPOINT = unix:///var/run/docker.sock
DOCKER_API_VERSION = v1.21
```

2. cd `humpback-agent` folder, run `go run main.go`
3. `curl http://localhost:8500/v1/dockerinfo` check the service status.

## License

Apache-2.0

## Changelog

[CHANGELOG.md](CHANGELOG.md)
