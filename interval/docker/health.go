package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/system"
	"humpback-agent/config"
	localTypes "humpback-agent/types"
)

func (d *DockerDriver) HealthDockerInfo() (*localTypes.DockerInfo, error) {
	dockerInfo, err := d.DockerInfo()
	if err != nil {
		return nil, err
	}
	dockerVersion, err := d.DockerVersion()
	if err != nil {
		return nil, err
	}
	return &localTypes.DockerInfo{
		Id:             dockerInfo.ID,
		Name:           dockerInfo.Name,
		Version:        dockerVersion.Version,
		APIVersion:     dockerVersion.APIVersion,
		MinAPIVersion:  dockerVersion.MinAPIVersion,
		DockerRootDir:  dockerInfo.DockerRootDir,
		StorageDriver:  dockerInfo.Driver,
		LoggingDriver:  dockerInfo.LoggingDriver,
		VolumePlugins:  dockerInfo.Plugins.Volume,
		NetworkPlugins: dockerInfo.Plugins.Network,
		NCPU:           dockerInfo.NCPU,
		MemTotal:       dockerInfo.MemTotal,
	}, nil
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
