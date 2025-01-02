package controller

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	v1model "humpback-agent/internal/api/v1/model"
)

type ContainerControllerInterface interface {
	BaseController() ControllerInterface
	Inspect(ctx context.Context, request *v1model.InspectContainerRequest) *v1model.ObjectResult
	List(ctx context.Context, request *v1model.QueryContainerRequest) *v1model.ObjectResult
	Create(ctx context.Context, request *v1model.CreateContainerRequest) *v1model.ObjectResult
	Update(ctx context.Context, request *v1model.UpdateContainerRequest) *v1model.ObjectResult
	Delete(ctx context.Context, request *v1model.DeleteContainerRequest) *v1model.ObjectResult
	Start(ctx context.Context, request *v1model.StartContainerRequest) *v1model.ObjectResult
	Restart(ctx context.Context, request *v1model.RestartContainerRequest) *v1model.ObjectResult
	Stop(ctx context.Context, request *v1model.StopContainerRequest) *v1model.ObjectResult
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

func (controller *ContainerController) Inspect(ctx context.Context, request *v1model.InspectContainerRequest) *v1model.ObjectResult {
	var containerBody types.ContainerJSON
	err := controller.baseController.WithTimeout(ctx, func(ctx context.Context) error {
		var err error
		containerBody, err = controller.client.ContainerInspect(ctx, request.ContainerId)
		return err
	})

	if err != nil {
		if errdefs.IsNotFound(err) {
			return v1model.ObjectNotFoundErrorResult(v1model.ContainerNotFoundCode, v1model.ContainerNotFoundMsg)
		}
		return v1model.ObjectInternalErrorResult(v1model.ServerInternalErrorCode, v1model.ServerInternalErrorMsg)
	}
	return v1model.ResultWithObject(containerBody)
}

func (controller *ContainerController) List(ctx context.Context, request *v1model.QueryContainerRequest) *v1model.ObjectResult {
	return nil
}

func (controller *ContainerController) Create(ctx context.Context, request *v1model.CreateContainerRequest) *v1model.ObjectResult {
	return nil
}

func (controller *ContainerController) Update(ctx context.Context, request *v1model.UpdateContainerRequest) *v1model.ObjectResult {
	return nil
}

func (controller *ContainerController) Delete(ctx context.Context, request *v1model.DeleteContainerRequest) *v1model.ObjectResult {
	return nil
}

func (controller *ContainerController) Restart(ctx context.Context, request *v1model.RestartContainerRequest) *v1model.ObjectResult {
	return nil
}

func (controller *ContainerController) Start(ctx context.Context, request *v1model.StartContainerRequest) *v1model.ObjectResult {
	return nil
}

func (controller *ContainerController) Stop(ctx context.Context, request *v1model.StopContainerRequest) *v1model.ObjectResult {
	return nil
}
