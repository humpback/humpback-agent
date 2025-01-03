package controller

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	v1model "humpback-agent/internal/api/v1/model"
	"humpback-agent/internal/utils"
)

type ContainerControllerInterface interface {
	BaseController() ControllerInterface
	Get(ctx context.Context, request *v1model.GetContainerRequest) *v1model.ObjectResult
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

func (controller *ContainerController) Get(ctx context.Context, request *v1model.GetContainerRequest) *v1model.ObjectResult {
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
	var containers []types.Container
	err := controller.baseController.WithTimeout(ctx, func(ctx context.Context) error {
		var err error
		filterArgs := filters.NewArgs()
		for key, value := range request.Filters {
			filterArgs.Add(key, value)
		}
		containers, err = controller.client.ContainerList(ctx, container.ListOptions{
			All:     request.All, // 是否包括已停止的容器
			Size:    request.Size,
			Latest:  request.Latest,
			Since:   request.Since,
			Before:  request.Before,
			Limit:   request.Limit,
			Filters: filterArgs,
		})
		return err
	})

	if err != nil {
		return v1model.ObjectInternalErrorResult(v1model.ServerInternalErrorCode, v1model.ServerInternalErrorMsg)
	}
	return v1model.ResultWithObject(containers)
}

func (controller *ContainerController) Create(ctx context.Context, request *v1model.CreateContainerRequest) *v1model.ObjectResult {
	// [1] 拉取镜像（如果需要）
	if request.AlwaysPull {
		if ret := controller.BaseController().Image().Pull(ctx, &v1model.PullImageRequest{Image: request.Image}); ret.Error != nil {
			return ret
		}
	}

	containerConfig := &container.Config{
		Image:      request.Image,
		Cmd:        request.Command,
		Entrypoint: request.Entrypoint,
		WorkingDir: request.WorkingDir,
		Tty:        request.TTY,
		OpenStdin:  request.Interactive,
		Env:        utils.MapToEnv(request.Env),
		Labels:     request.Labels,
	}

	hostConfig := &container.HostConfig{
		AutoRemove:      request.AutoRemove,
		PortBindings:    utils.MapPorts(request.PortMap),
		PublishAllPorts: request.PublishAll,
		LogConfig:       container.LogConfig{
			//Type:   request.Logger.Driver,
			//Config: request.Logger.Options,
		},
		//Binds:       utils.MapToBinds(request.),
		//NetworkMode: container.NetworkMode(request.Network.Mode),
		//RestartPolicy: container.RestartPolicy{
		//	Name: request.RestartPolicy,
		//},
		//Privileged: request.Runtime.Privileged,
		//Init:       &request.Runtime.Init,
		//Runtime:    request.Runtime.Runtime,
		//Sysctls:    request.Sysctls,
		//ShmSize:    request.ShmSize,
		//Resources: container.Resources{
		//	Memory:            request.MemoryLimit,
		//	MemoryReservation: request.MemoryReserve,
		//	CPUQuota:          request.CPUQuota,
		//},
		//CapAdd:  request.CapAdd,
		//CapDrop: request.CapDrop,
	}

	// [4] 配置网络选项
	networkConfig := &network.NetworkingConfig{
		//EndpointsConfig: map[string]*network.EndpointSettings{
		//	request.NetworkMode: {
		//		IPAMConfig: &network.EndpointIPAMConfig{
		//			IPv4Address: request.IPv4Address,
		//			IPv6Address: request.IPv6Address,
		//		},
		//		MacAddress: request.MacAddress,
		//	},
		//},
	}

	var containerInfo container.CreateResponse
	err := controller.baseController.WithTimeout(ctx, func(ctx context.Context) error {
		var err error
		containerInfo, err = controller.client.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, nil, request.Name)
		if err != nil {
			return err
		}

		//启动容器
		if !request.OnlyCreate {
			return controller.client.ContainerStart(ctx, containerInfo.ID, container.StartOptions{})
		}
		return nil
	})

	if err != nil {
		if errdefs.IsConflict(err) {
			return v1model.ObjectInternalErrorResult(v1model.ContainerCreateConflictErrorCode, v1model.ContainerCreateConflictErrorMsg)
		}
		return v1model.ObjectInternalErrorResult(v1model.ContainerCreateErrorCode, v1model.ContainerCreateErrorMsg)
	}
	return v1model.ResultWithObject(containerInfo.ID)
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
