package controller

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	v1model "humpback-agent/api/v1/model"
	"humpback-agent/internal/schedule"
	"humpback-agent/model"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
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
	Stats(ctx context.Context, request *v1model.GetContainerStatsRequest) *v1model.ObjectResult
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
			return v1model.ObjectNotFoundErrorResult(v1model.ContainerNotFoundCode, err.Error())
		}
		return v1model.ObjectInternalErrorResult(v1model.ContainerGetErrorCode, err.Error())
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
	result := controller.createInternal(ctx, request)

	if result.Error != nil {
		containerMeta := model.ContainerMeta{
			ContainerName: request.ContainerName,
			State:         model.ContainerStatusFailed,
			ErrorMsg:      result.Error.ErrMsg,
		}
		controller.BaseController().FailureChan() <- containerMeta
	}

	return result
}

func (controller *ContainerController) createInternal(ctx context.Context, request *v1model.CreateContainerRequest) *v1model.ObjectResult {

	value, _ := json.MarshalIndent(request, "", "    ")
	fmt.Printf("%s\n", value)

	isJob := false

	if request.Labels == nil {
		request.Labels = make(map[string]string)
	}

	request.Labels[v1model.ContainerLabelServiceId] = request.ServiceId
	request.Labels[v1model.ContainerLabelGroupId] = request.GroupId
	request.Labels[v1model.ContainerLabelServiceName] = request.ServiceName

	if request.ScheduleInfo != nil && len(request.ScheduleInfo.Rules) > 0 {
		isJob = true
		var jobRules string
		if len(request.ScheduleInfo.Rules) > 0 {
			jobRules = strings.Join(request.ScheduleInfo.Rules, ";")
		}

		request.Labels[schedule.HumpbackJobRulesLabel] = jobRules
		request.Labels[schedule.HumpbackJobAlwaysPullLabel] = strconv.FormatBool(request.AlwaysPull)
		request.Labels[schedule.HumpbackJobMaxTimeoutLabel] = request.ScheduleInfo.Timeout
	}

	containerConfig := &container.Config{
		Image:  request.Image,
		Env:    request.Envs,
		Labels: request.Labels,
	}

	if request.Command != "" {
		containerConfig.Cmd = strings.Fields(request.Command)
	}

	hostConfig := &container.HostConfig{
		Privileged: request.Privileged,
	}

	if request.Capabilities != nil {
		capAdd := request.Capabilities.CapAdd
		if len(capAdd) > 0 {
			hostConfig.CapAdd = capAdd
		}

		capDrop := request.Capabilities.CapDrop
		if len(capDrop) > 0 {
			hostConfig.CapDrop = capDrop
		}
	}

	if request.LogConfig != nil {
		hostConfig.LogConfig = container.LogConfig{
			Type:   request.LogConfig.Type,
			Config: request.LogConfig.Config,
		}
	}

	if request.Resources != nil {
		hostConfig.Resources = container.Resources{}
		if request.Resources.Memory > 0 {
			mLimit := int64(request.Resources.Memory * 1024 * 1024)
			if mLimit < 6*1024*1024 {
				mLimit = 6 * 1024 * 1024
			}
			hostConfig.Resources.Memory = mLimit
		}
		if request.Resources.MemoryReservation > 0 {
			hostConfig.Resources.MemoryReservation = int64(request.Resources.MemoryReservation * 1024 * 1024)
		}
		if hostConfig.Resources.Memory < hostConfig.Resources.MemoryReservation && hostConfig.Resources.Memory != 0 {
			hostConfig.Resources.Memory = hostConfig.Resources.MemoryReservation
		}
		if request.Resources.MaxCpuUsage > 0 {
			cpuLimit := int64(request.Resources.MaxCpuUsage * 1000000000 / 100)
			hostConfig.Resources.NanoCPUs = cpuLimit
		}
	}

	if request.RestartPolicy != nil {
		restartPolicyModeName := request.RestartPolicy.Mode
		maxRetryCount := request.RestartPolicy.MaxRetryCount
		if isJob { //定时任务强制设置为No
			restartPolicyModeName = v1model.RestartPolicyModeNo
			maxRetryCount = 0
		}
		hostConfig.RestartPolicy = container.RestartPolicy{
			Name: container.RestartPolicyMode(restartPolicyModeName),
		}

		if restartPolicyModeName == v1model.RestartPolicyModeOnFail {
			hostConfig.RestartPolicy.MaximumRetryCount = maxRetryCount
		}
	}

	var networkConfig *network.NetworkingConfig
	if request.Network != nil {
		hostname := request.Network.Hostname
		if request.Network.UseMachineHostname {
			hostname, _ = os.Hostname()
		}
		if request.Network.Mode == v1model.NetworkModeCustom { //构建自定义网络
			containerConfig.Hostname = hostname
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
			containerConfig.Hostname = hostname
			networkConfig = &network.NetworkingConfig{
				EndpointsConfig: map[string]*network.EndpointSettings{
					"bridge": {
						NetworkID: "bridge",
					},
				},
			}
		}

		portBindings := nat.PortMap{}
		if len(request.Network.Ports) > 0 {
			containerConfig.ExposedPorts = make(nat.PortSet)
			for _, bindPort := range request.Network.Ports {
				proto := strings.ToLower(bindPort.Protocol)
				if proto != "tcp" && proto != "udp" {
					proto = "tcp" // 默认使用 TCP
				}
				port, err := nat.NewPort(proto, strconv.Itoa(int(bindPort.ContainerPort)))
				if err != nil {
					return v1model.ObjectInternalErrorResult(v1model.ContainerCreateErrorCode, err.Error())
				}
				hostPort := int(bindPort.HostPort)
				if hostPort == 0 {
					if hostPort, err = controller.BaseController().AllocPort(proto); err != nil {
						return v1model.ObjectInternalErrorResult(v1model.ContainerCreateErrorCode, err.Error())
					}
				}
				containerConfig.ExposedPorts[port] = struct{}{}
				portBindings[port] = []nat.PortBinding{{HostPort: strconv.Itoa(hostPort)}}
			}
			hostConfig.PortBindings = portBindings
		} else {
			hostConfig.PublishAllPorts = true //若请求中没设置端口, 则自动暴露镜像Dockerfile中的所有端口
		}
	}

	//处理卷配置绑定
	if err := controller.buildHostConfigVolumesWithRequest(request.Volumes, hostConfig); err != nil {
		return v1model.ObjectInternalErrorResult(v1model.ContainerCreateErrorCode, err.Error())
	}

	//先尝试处理镜像
	if pullResult := controller.BaseController().Image().AttemptPull(context.Background(), request.Image, request.AlwaysPull); pullResult.Error != nil {
		return v1model.ObjectInternalErrorResult(v1model.ImagePullErrorCode, pullResult.Error.ErrMsg)
	}

	var containerInfo container.CreateResponse
	err := controller.baseController.WithTimeout(ctx, func(ctx context.Context) error {
		var createdErr error
		containerInfo, createdErr = controller.client.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, nil, request.ContainerName)
		if createdErr != nil {
			return createdErr
		}
		if !isJob { //Job容器创建后自动启动
			return controller.client.ContainerStart(ctx, containerInfo.ID, container.StartOptions{})
		}
		return nil
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

	var (
		logs = make([]string, 0)
		line = 0
	)

	if err := controller.baseController.WithTimeout(ctx, func(ctx context.Context) error {
		//获取日志流
		logReader, logsErr := controller.client.ContainerLogs(ctx, request.ContainerId, options)
		if logsErr != nil {
			return logsErr
		}

		defer logReader.Close()
		hdr := make([]byte, 8)
		for {
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
			if line > 10000 {
				return errors.New("The maximum limit of 10,000 rows is exceeded.")
			}

			logs = append(logs, string(dat))
			line++
		}
	}); err != nil {
		return v1model.ObjectInternalErrorResult(v1model.ContainerLogsErrorCode, err.Error())
	}
	return v1model.ResultWithObject(logs)
}

func (controller *ContainerController) Stats(ctx context.Context, request *v1model.GetContainerStatsRequest) *v1model.ObjectResult {
	containerStats := container.StatsResponse{}
	if err := controller.baseController.WithTimeout(ctx, func(ctx context.Context) error {
		statsReader, statsErr := controller.client.ContainerStats(ctx, request.ContainerId, true)
		if statsErr != nil {
			return statsErr
		}
		defer statsReader.Body.Close()
		return json.NewDecoder(statsReader.Body).Decode(&containerStats)
	}); err != nil {
		return v1model.ObjectInternalErrorResult(v1model.ContainerStatsErrorCode, err.Error())
	}
	return v1model.ResultWithObject(model.ParseContainerStats(&containerStats))
}

func (controller *ContainerController) buildHostConfigVolumesWithRequest(reqVolumes []*v1model.ServiceVolume, hostConfig *container.HostConfig) error {
	configNames := controller.BaseController().GetConfigNamesWithVolumes(reqVolumes)
	configPaths, err := controller.BaseController().BuildVolumesWithConfigNames(configNames)
	if err != nil {
		return err
	}

	var mounts []mount.Mount
	for _, volume := range reqVolumes {
		if volume.Type == v1model.ServiceVolumeTypeBind {
			matches := re.FindStringSubmatch(volume.Source)
			if len(matches) > 1 {
				path, ret := configPaths[matches[1]]
				if !ret {
					return fmt.Errorf("invalid %s volume path: %s", volume.Type, volume.Source)
				}
				volume.Source = path
			}

			// 将配置转换为 mount.Mount
			mounts = append(mounts, mount.Mount{
				Type:     mount.Type(volume.Type),
				Source:   volume.Source,
				Target:   volume.Target,
				ReadOnly: volume.Readonly,
			})
		}
	}
	hostConfig.Mounts = mounts
	return nil
}
