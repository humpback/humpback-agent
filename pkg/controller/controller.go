package controller

import (
	"context"
	"github.com/docker/docker/client"
	"humpback-agent/pkg/model"
	"time"
)

type InternalController interface {
	WithTimeout(ctx context.Context, callback func(context.Context) error) error
	DockerEngine(ctx context.Context) (*model.DockerEngineInfo, error)
}

type ControllerInterface interface {
	InternalController
	Image() ImageControllerInterface
	Container() ContainerControllerInterface
}

type BaseController struct {
	client     *client.Client
	reqTimeout time.Duration
	image      ImageControllerInterface
	container  ContainerControllerInterface
}

func NewController(client *client.Client, reqTimeout time.Duration) ControllerInterface {
	baseController := &BaseController{
		client:     client,
		reqTimeout: reqTimeout,
	}

	baseController.image = NewImageController(baseController, client)
	baseController.container = NewContainerController(baseController, client)
	return baseController
}

func (controller *BaseController) WithTimeout(ctx context.Context, callback func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, controller.reqTimeout)
	defer cancel()
	return callback(ctx)
}

func (controller *BaseController) DockerEngine(ctx context.Context) (*model.DockerEngineInfo, error) {
	var engineInfo model.DockerEngineInfo
	serverVersion, getErr := controller.client.ServerVersion(ctx)
	if getErr != nil {
		return nil, getErr
	}

	dockerInfo, infoErr := controller.client.Info(ctx)
	if infoErr != nil {
		return nil, infoErr
	}

	engineInfo.Version = dockerInfo.ServerVersion
	engineInfo.APIVersion = serverVersion.APIVersion
	engineInfo.RootDirectory = dockerInfo.DockerRootDir
	engineInfo.StorageDriver = dockerInfo.Driver
	engineInfo.LoggingDriver = dockerInfo.LoggingDriver
	engineInfo.VolumePlugins = dockerInfo.Plugins.Volume
	engineInfo.NetworkPlugins = dockerInfo.Plugins.Network
	return &engineInfo, nil
}

func (controller *BaseController) Image() ImageControllerInterface {
	return controller.image
}

func (controller *BaseController) Container() ContainerControllerInterface {
	return controller.container
}
