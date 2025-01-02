package controller

import (
	"github.com/docker/docker/client"
)

type InternalController interface{}

type ControllerInterface interface {
	InternalController
	Image() ImageControllerInterface
	Container() ContainerControllerInterface
}

type BaseController struct {
	client    *client.Client
	image     ImageControllerInterface
	container ContainerControllerInterface
}

func NewController(client *client.Client) ControllerInterface {
	baseController := &BaseController{
		client: client,
	}

	baseController.image = NewImageController(baseController, client)
	baseController.container = NewContainerController(baseController, client)
	return baseController
}

func (controller *BaseController) Image() ImageControllerInterface {
	return controller.image
}

func (controller *BaseController) Container() ContainerControllerInterface {
	return controller.container
}
