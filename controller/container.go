package controller

import (
	"context"
	"encoding/binary"
	v1model "humpback-agent/api/v1/model"
	"humpback-agent/model"
	"io"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
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
	Logs(ctx context.Context, request *v1model.GetContainerLogsRequest) *v1model.ObjectResult
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
		return v1model.ObjectInternalErrorResult(v1model.ServerInternalErrorCode, err.Error())
	}
	return v1model.ResultWithObject(containerBody)
}

func (controller *ContainerController) List(ctx context.Context, request *v1model.QueryContainerRequest) *v1model.ObjectResult {
	var containers []types.Container
	err := controller.baseController.WithTimeout(ctx, func(ctx context.Context) error {
		filterArgs := filters.NewArgs()
		for key, value := range request.Filters {
			filterArgs.Add(key, value)
		}
		var queryErr error
		containers, queryErr = controller.client.ContainerList(ctx, container.ListOptions{
			All:     request.All, // 是否包括已停止的容器
			Size:    request.Size,
			Latest:  request.Latest,
			Since:   request.Since,
			Before:  request.Before,
			Limit:   request.Limit,
			Filters: filterArgs,
		})
		return queryErr
	})

	if err != nil {
		return v1model.ObjectInternalErrorResult(v1model.ServerInternalErrorCode, v1model.ServerInternalErrorMsg)
	}
	return v1model.ResultWithObject(containers)
}

func (controller *ContainerController) Create(ctx context.Context, request *v1model.CreateContainerRequest) *v1model.ObjectResult {
	if request.AlwaysPull { //拉取镜像（如果需要）
		if ret := controller.BaseController().Image().Pull(ctx, &v1model.PullImageRequest{Image: request.Image}); ret.Error != nil {
			return ret
		}
	}

	containerConfig := &container.Config{
		Image:  request.Image,
		Env:    request.Envs,
		Labels: request.Labels,
	}

	if request.Command != "" {
		containerConfig.Cmd = strslice.StrSlice{request.Command}
	}

	hostConfig := &container.HostConfig{}
	if request.RestartPolicy != nil {
		hostConfig.RestartPolicy = container.RestartPolicy{
			Name:              container.RestartPolicyMode(request.RestartPolicy.Mode),
			MaximumRetryCount: request.RestartPolicy.MaxRetryCount,
		}
	}

	var networkConfig *network.NetworkingConfig
	if request.Network != nil {
		if request.Network.Mode == v1model.NetworkModeCustom { //构建自定义网络
			containerConfig.Hostname = request.Network.Hostname
			if request.Network.NetworkName != "" {
				networkResult := controller.BaseController().Network().Create(ctx, &v1model.CreateNetworkRequest{NetworkName: request.Network.NetworkName, Driver: "bridge", Scope: "local"})
				if networkResult.Error != nil {
					return networkResult
				}
				hostConfig.NetworkMode = container.NetworkMode(request.Network.NetworkName)
				networkConfig = &network.NetworkingConfig{
					EndpointsConfig: map[string]*network.EndpointSettings{
						request.Network.NetworkName: {
							NetworkID: networkResult.ObjectId,
						},
					},
				}
			}
		} else if request.Network.Mode == v1model.NetworkModeHost {
			hostConfig.NetworkMode = container.NetworkMode(request.Network.Mode)
			hostConfig.PublishAllPorts = true
		} else if request.Network.Mode == v1model.NetworkModeBridge { // 桥接, 配置 PortBindings
			hostConfig.NetworkMode = container.NetworkMode(request.Network.Mode)
			containerConfig.Hostname = request.Network.Hostname
			portBindings := nat.PortMap{}
			for _, bindPort := range request.Network.Ports {
				proto := strings.ToLower(bindPort.Protocol)
				if proto != "tcp" && proto != "udp" {
					proto = "tcp" // 默认使用 TCP
				}
				port, err := nat.NewPort(proto, strconv.Itoa(int(bindPort.ContainerPort)))
				if err != nil {
					return v1model.ObjectInternalErrorResult(v1model.ContainerCreateErrorCode, err.Error())
				}
				portBindings[port] = []nat.PortBinding{{HostPort: strconv.Itoa(int(bindPort.HostPort))}}
			}
			hostConfig.PortBindings = portBindings
			networkConfig = &network.NetworkingConfig{
				EndpointsConfig: map[string]*network.EndpointSettings{
					"bridge": {
						NetworkID: "bridge",
					},
				},
			}
		}
	}

	var containerInfo container.CreateResponse
	err := controller.baseController.WithTimeout(ctx, func(ctx context.Context) error {
		var createdErr error
		containerInfo, createdErr = controller.client.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, nil, request.ContainerName)
		if createdErr != nil {
			return createdErr
		}
		return controller.client.ContainerStart(ctx, containerInfo.ID, container.StartOptions{})
	})

	if err != nil {
		return v1model.ObjectInternalErrorResult(v1model.ContainerCreateErrorCode, err.Error())
	}
	return v1model.ResultWithObjectId(containerInfo.ID)
}

func (controller *ContainerController) Update(ctx context.Context, request *v1model.UpdateContainerRequest) *v1model.ObjectResult {
	return nil
}

func (controller *ContainerController) Delete(ctx context.Context, request *v1model.DeleteContainerRequest) *v1model.ObjectResult {
	var containerId string
	if err := controller.baseController.WithTimeout(ctx, func(ctx context.Context) error {
		containerBody, inspectErr := controller.client.ContainerInspect(ctx, request.ContainerId)
		if inspectErr != nil {

			return inspectErr
		}
		containerId = containerBody.ID
		return controller.client.ContainerRemove(ctx, request.ContainerId, container.RemoveOptions{Force: request.Force})
	}); err != nil {
		return v1model.ObjectInternalErrorResult(v1model.ContainerDeleteErrorCode, err.Error())
	}
	return v1model.ResultWithObjectId(containerId)
}

func (controller *ContainerController) Restart(ctx context.Context, request *v1model.RestartContainerRequest) *v1model.ObjectResult {
	var containerId string
	if err := controller.baseController.WithTimeout(ctx, func(ctx context.Context) error {
		containerBody, inspectErr := controller.client.ContainerInspect(ctx, request.ContainerId)
		if inspectErr != nil {
			return inspectErr
		}
		containerId = containerBody.ID
		return controller.client.ContainerRestart(ctx, request.ContainerId, container.StopOptions{})
	}); err != nil {
		return v1model.ObjectInternalErrorResult(v1model.ContainerDeleteErrorCode, err.Error())
	}
	return v1model.ResultWithObjectId(containerId)
}

func (controller *ContainerController) Start(ctx context.Context, request *v1model.StartContainerRequest) *v1model.ObjectResult {
	var containerId string
	if err := controller.baseController.WithTimeout(ctx, func(ctx context.Context) error {
		containerBody, inspectErr := controller.client.ContainerInspect(ctx, request.ContainerId)
		if inspectErr != nil {
			return inspectErr
		}
		containerId = containerBody.ID
		return controller.client.ContainerStart(ctx, request.ContainerId, container.StartOptions{})
	}); err != nil {
		return v1model.ObjectInternalErrorResult(v1model.ContainerDeleteErrorCode, err.Error())
	}
	return v1model.ResultWithObjectId(containerId)
}

func (controller *ContainerController) Stop(ctx context.Context, request *v1model.StopContainerRequest) *v1model.ObjectResult {
	var containerId string
	if err := controller.baseController.WithTimeout(ctx, func(ctx context.Context) error {
		containerBody, inspectErr := controller.client.ContainerInspect(ctx, request.ContainerId)
		if inspectErr != nil {
			return inspectErr
		}
		containerId = containerBody.ID
		return controller.client.ContainerStop(ctx, request.ContainerId, container.StopOptions{})
	}); err != nil {
		return v1model.ObjectInternalErrorResult(v1model.ContainerDeleteErrorCode, err.Error())
	}
	return v1model.ResultWithObjectId(containerId)
}

func (controller *ContainerController) Logs(ctx context.Context, request *v1model.GetContainerLogsRequest) *v1model.ObjectResult {
	options := container.LogsOptions{
		ShowStdout: true, // 显示标准输出
		ShowStderr: true, // 显示标准错误
	}

	if request.Follow != nil {
		options.Follow = *request.Follow
	}

	if request.Tail != nil {
		options.Tail = *request.Tail
	}

	if request.Since != nil {
		options.Since = *request.Since
	}

	if request.Until != nil {
		options.Until = *request.Until
	}

	if request.Timestamps != nil {
		options.Timestamps = *request.Timestamps
	}

	if request.Details != nil {
		options.Details = *request.Details
	}

	dockerLogs := model.DockerContainerLog{
		DockerLogs: []model.DockerLog{},
	}

	if err := controller.baseController.WithTimeout(ctx, func(ctx context.Context) error {
		containerBody, inspectErr := controller.client.ContainerInspect(ctx, request.ContainerId)
		if inspectErr != nil {
			return inspectErr
		}

		dockerLogs.ContainerId = containerBody.ID
		//获取日志流
		logReader, logsErr := controller.client.ContainerLogs(ctx, request.ContainerId, options)
		if logsErr != nil {
			return logsErr
		}

		defer logReader.Close()
		hdr := make([]byte, 8)
		for {
			var docLog model.DockerLog
			_, readErr := logReader.Read(hdr)
			if readErr != nil {
				if readErr == io.EOF {
					return nil
				}
				return readErr
			}

			count := binary.BigEndian.Uint32(hdr[4:])
			dat := make([]byte, count)
			_, readErr = logReader.Read(dat)
			if readErr != nil && readErr != io.EOF {
				return readErr
			}

			time, log, found := strings.Cut(string(dat), " ")
			if found {
				docLog.Time = time
				docLog.Log = log
				switch hdr[0] {
				case 1:
					docLog.Stream = "Stdout"
				default:
					docLog.Stream = "Stderr"
				}
				dockerLogs.DockerLogs = append(dockerLogs.DockerLogs, docLog)
			}
		}
	}); err != nil {
		return v1model.ObjectInternalErrorResult(v1model.ContainerLogsErrorCode, err.Error())
	}
	return v1model.ResultWithObject(dockerLogs)
}
