package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/system"
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

func (d *DockerDriver) DockerInfo() (system.Info, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.DockerArgs().Timeout)
	defer cancel()
	return d.client.Info(ctx)
}

func (d *DockerDriver) DockerVersion() (types.Version, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.DockerArgs().Timeout)
	defer cancel()
	return d.client.ServerVersion(ctx)
}
