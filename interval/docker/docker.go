package docker

import (
	"github.com/docker/docker/client"
	"humpback-agent/config"
)

type DockerDriver struct {
	client *client.Client
}

func NewDockerDriver() (*DockerDriver, error) {
	dockerConfig := config.DockerArgs()
	client, err := client.NewClientWithOpts(client.WithHost(dockerConfig.Host), client.WithVersion(dockerConfig.ApiVersion))
	if err != nil {
		return nil, err
	}
	return &DockerDriver{client: client}, nil
}

func (d *DockerDriver) Close() error {
	return d.client.Close()
}
