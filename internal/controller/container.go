package controller

import (
	"github.com/docker/docker/client"
	v1model "humpback-agent/internal/api/v1/model"
)

type ContainerControllerInterface interface {
	BaseController() ControllerInterface
	Get(request *v1model.GetContainerRequest) *v1model.ObjectResult
	List(request *v1model.QueryContainerRequest) *v1model.ObjectResult
	Create(request *v1model.CreateContainerRequest) *v1model.ObjectResult
	Update(request *v1model.UpdateContainerRequest) *v1model.ObjectResult
	Delete(request *v1model.DeleteContainerRequest) *v1model.ObjectResult
}

type ContainerController struct {
	baseController ControllerInterface
	client         *client.Client
}

func NewContainerController(baseController ControllerInterface, client *client.Client) ContainerControllerInterface {
	return &ContainerController{
		baseController: baseController,
		client:         client,
	}
}

func (controller *ContainerController) BaseController() ControllerInterface {
	return controller.baseController
}

func (controller *ContainerController) Get(request *v1model.GetContainerRequest) *v1model.ObjectResult {
	return nil
}

func (controller *ContainerController) List(request *v1model.QueryContainerRequest) *v1model.ObjectResult {
	return nil
}

func (controller *ContainerController) Create(request *v1model.CreateContainerRequest) *v1model.ObjectResult {
	return nil
}

func (controller *ContainerController) Update(request *v1model.UpdateContainerRequest) *v1model.ObjectResult {
	return nil
}

func (controller *ContainerController) Delete(request *v1model.DeleteContainerRequest) *v1model.ObjectResult {
	return nil
}
