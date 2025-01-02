package controller

import (
	"context"
	"github.com/docker/docker/client"
	"time"
)

type InternalController interface {
	WithTimeout(ctx context.Context, callback func(context.Context) error) error
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
		client: client,
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

func (controller *BaseController) Image() ImageControllerInterface {
	return controller.image
}

func (controller *BaseController) Container() ContainerControllerInterface {
	return controller.container
}
