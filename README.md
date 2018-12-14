## humpback-agent

Humpback Agent 主要为 [Humpback](https://github.com/humpback/humpback) 平台提供容器访问功能。[API 文档](https://github.com/humpback/humpback-agent/wiki)

## How to develop

### #Prepare

1. First, you need an server that installed and already running docker(recommend Linux).
2. `[Optional]` If you are develop on windows, you should expose docker host with tcp protocol(default is `unix:///var/run/docker.sock`). How to do it?

```bash
# convert unix-connect as container port 6666, mapping the container port 6666 as host port 1234
docker run -d --restart=always \
    -p 0.0.0.0:1234:6666 \
    -v /var/run/docker.sock:/var/run/docker.sock \
    alpine/socat \
    tcp-listen:6666,fork,reuseaddr unix-connect:/var/run/docker.sock
```

expose your docker host at `tcp:<ip>:1234` [more info](https://hub.docker.com/r/alpine/socat/).

### #Init dependencies

1. Use `go get` install the dependencies.

**Notice: GOPATH is very important!**

### #Build & Run

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
