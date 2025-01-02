package app

import (
	"github.com/docker/docker/client"
	"humpback-agent/internal/config"
)

func buildDockerClient(config *config.DockerConfig) (*client.Client, error) {
	opts := []client.Opt{
		client.WithHost(config.Host),
		//client.WithHTTPClient(&http.Client{
		//	Timeout: config.DockerTimeoutOpts.Connection,
		//}),
	}

	if config.DockerTLSOpts.Enabled {
		opts = append(opts, client.WithTLSClientConfig(config.DockerTLSOpts.CAPath, config.DockerTLSOpts.CertPath, config.DockerTLSOpts.KeyPath))
	}

	if config.AutoNegotiate {
		opts = append(opts, client.WithAPIVersionNegotiation()) // Humpback 使用 docker sdk 与 docker daemon 自动协商版本
	} else {
		opts = append(opts, client.WithVersion(config.Version))
	}
	return client.NewClientWithOpts(opts...)
}
