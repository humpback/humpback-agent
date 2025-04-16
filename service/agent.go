package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"humpback-agent/api"
	v1model "humpback-agent/api/v1/model"
	"humpback-agent/config"
	"humpback-agent/controller"
	reqclient "humpback-agent/internal/client"
	"humpback-agent/internal/docker"
	"humpback-agent/internal/schedule"
	"humpback-agent/model"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

type AgentService struct {
	sync.RWMutex
	config            *config.AppConfig
	apiServer         *api.APIServer
	httpClient        *http.Client
	scheduler         schedule.TaskSchedulerInterface
	controller        controller.ControllerInterface
	failureChan       chan model.ContainerMeta
	containers        map[string]*model.ContainerInfo
	failureContainers map[string]*model.ContainerInfo
}

func NewAgentService(ctx context.Context, config *config.AppConfig) (*AgentService, error) {
	//构建Agent服务
	agentService := &AgentService{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Health.Timeout,
		},
		containers:        make(map[string]*model.ContainerInfo),
		failureContainers: make(map[string]*model.ContainerInfo),
		failureChan:       make(chan model.ContainerMeta, 10),
	}
	//构建Docker-client
	dockerClient, err := docker.BuildDockerClient(config.DockerConfig)
	if err != nil {
		return nil, err
	}

	//构建API和Controller接口
	appController := controller.NewController(
		dockerClient,
		agentService.sendConfigValuesRequest,
		config.VolumesConfig.RootDirectory,
		config.DockerTimeoutOpts.Request,
		agentService.failureChan,
	)

	apiServer, err := api.NewAPIServer(appController, config.APIConfig)
	if err != nil {
		return nil, err
	}

	agentService.apiServer = apiServer
	agentService.scheduler = schedule.NewJobScheduler(dockerClient) //构建任务定时调度器
	agentService.controller = appController

	//启动先加载本地所有容器
	if err = agentService.loadDockerContainers(ctx); err != nil {
		return nil, err
	}

	//启动服务API
	if err = apiServer.Startup(ctx); err != nil {
		return nil, err
	}

	//启动心跳
	go agentService.heartbeatLoop()

	go agentService.watchMetaChange()

	//启动docker事件监听
	go agentService.watchDockerEvents(ctx, dockerClient)
	//启动定时任务调度器
	agentService.scheduler.Start()
	//初始化一次所有定时容器, 加入调度器
	for _, container := range agentService.containers {
		if _, ret := container.Labels[schedule.HumpbackJobRulesLabel]; ret { //job定时容器, 交给定时调度器
			agentService.addToScheduler(container.ContainerId, container.ContainerName, container.Image, container.Labels)
		}
	}

	return agentService, nil
}

func (agentService *AgentService) Shutdown(ctx context.Context) {
	if agentService.apiServer != nil {
		if err := agentService.apiServer.Stop(ctx); err != nil {
			logrus.Errorf("Humpback Agent api server stop error, %s", err.Error())
		}
	}
	//关闭定时任务调度器
	agentService.scheduler.Stop()
}

func (agentService *AgentService) loadDockerContainers(ctx context.Context) error {
	result := agentService.controller.Container().List(ctx, &v1model.QueryContainerRequest{All: true})
	if result.Error != nil {
		return fmt.Errorf("load container list error")
	}

	containers := result.Object.([]types.Container)
	slog.Info("[loadDockerContainers] contianer len.", "Length", len(containers))
	agentService.Lock()
	for _, container := range containers {
		result = agentService.controller.Container().Get(ctx, &v1model.GetContainerRequest{ContainerId: container.ID})
		if result.Error != nil {
			return fmt.Errorf("load container inspect %s error, %v", container.ID, result.Error)
		}
		containerInfo := model.ParseContainerInfo(result.Object.(types.ContainerJSON))
		agentService.containers[container.ID] = containerInfo
		slog.Info("[loadDockerContainers] add to cache.", "ContainerID", container.ID, "Name", containerInfo.ContainerName, "Labels", containerInfo.Labels)
	}
	agentService.Unlock()
	return nil
}

func (agentService *AgentService) fetchContainer(ctx context.Context, containerId string) (*model.ContainerInfo, error) {
	result := agentService.controller.Container().Get(ctx, &v1model.GetContainerRequest{ContainerId: containerId})
	if result.Error != nil {
		return nil, fmt.Errorf("get container %s error, %v", containerId, result.Error)
	}
	return model.ParseContainerInfo(result.Object.(types.ContainerJSON)), nil
}

func (agentService *AgentService) heartbeatLoop() {
	for {
		if err := agentService.sendHealthRequest(context.Background()); err != nil {
			logrus.Errorf("heartbeat health send request error: %+v", err.Error())
		} else {
			logrus.Debugf("heartbeat health send request done at %s\n", time.Now().String())
		}
		time.Sleep(agentService.config.Health.Interval)
	}
}

func (agentService *AgentService) watchDockerEvents(ctx context.Context, dockerClient *client.Client) {
	eventChan, errChan := dockerClient.Events(ctx, types.EventsOptions{})
	for {
		select {
		case event := <-eventChan:
			agentService.handleDockerEvent(event)
		case err := <-errChan:
			logrus.Errorf("Watch docker event error, %v", err)
			time.Sleep(time.Second * 1)
		}
	}
}

func (agentService *AgentService) handleDockerEvent(message events.Message) {
	if message.Type == "container" {
		switch message.Action {
		case "create", "start", "stop", "die", "kill", "healthy", "unhealthy":
			containerInfo, err := agentService.fetchContainer(context.Background(), message.Actor.ID)
			if err != nil {
				logrus.Errorf("Docker create container %s event, %v", message.Actor.ID, err)
			}
			if containerInfo != nil {

				needReport := true

				if message.Action == "create" {
					if _, ret := containerInfo.Labels[schedule.HumpbackJobRulesLabel]; ret { //创建了一个job定时容器, 交给定时调度器
						agentService.addToScheduler(containerInfo.ContainerId, containerInfo.ContainerName, containerInfo.Image, containerInfo.Labels)
						//找到相同name的删除, 因为reCreate原因, 缓存先同步删除
						agentService.Lock()
						for containerId, container := range agentService.containers {
							if container.ContainerName == containerInfo.ContainerName {
								delete(agentService.containers, containerId)
								break
							}
						}
						agentService.Unlock()
					} else {
						// 非定时容器的create，不需要汇报心跳
						needReport = false
					}
				}

				agentService.Lock()
				old, ok := agentService.containers[containerInfo.ContainerId]
				if ok && old.State == containerInfo.State {
					needReport = false
				}
				agentService.containers[containerInfo.ContainerId] = containerInfo

				agentService.Unlock()

				if needReport {
					slog.Info("send heartbeat", "container", containerInfo.ContainerName, "action", message.Action, "status", containerInfo.State)
					agentService.sendHealthRequest(context.Background())
				}
			}
		case "destroy", "remove", "delete":
			if message.Action == "destroy" { //从job定时调度器删除, 无论是否在调度器中, 会自动处理
				agentService.removeFromScheduler(message.Actor.ID)
			}
			state := "unknow"
			//修改容器状态
			agentService.Lock()
			if containerInfo, ret := agentService.containers[message.Actor.ID]; ret {
				containerInfo.State = model.ContainerStatusRemoved
				state = model.ContainerStatusRemoved
			}
			agentService.Unlock()
			//主动通知一次心跳
			slog.Info("send heartbeat", "container", message.Actor.ID, "action", message.Action, "status", state)
			agentService.sendHealthRequest(context.Background())
			//缓存删除容器
			agentService.Lock()
			delete(agentService.containers, message.Actor.ID)
			agentService.Unlock()
		}
	}
}

func (agentService *AgentService) addToScheduler(containerId string, containerName string, containerImage string, containerLabels map[string]string) error {
	if value, ret := containerLabels[schedule.HumpbackJobRulesLabel]; ret && value != "Manual" {
		var (
			err        error
			timeout    time.Duration
			alwaysPull bool
			authStr    string
		)
		rules := strings.Split(value, ";")
		if value, ret = containerLabels[schedule.HumpbackJobMaxTimeoutLabel]; ret && value != "" {
			if timeout, err = time.ParseDuration(value); err != nil {
				return err
			}
		}

		if value, ret = containerLabels[schedule.HumpbackJobAlwaysPullLabel]; ret && value != "" {
			if alwaysPull, err = strconv.ParseBool(value); err != nil {
				return err
			}
		}

		if value, ret = containerLabels[schedule.HumpbackJobImageAuth]; ret {
			authStr = value
		}

		return agentService.scheduler.AddContainer(containerId, containerName, containerImage, alwaysPull, rules, authStr, timeout)
	}
	return nil
}

func (agentService *AgentService) removeFromScheduler(containerId string) error {
	return agentService.scheduler.RemoveContainer(containerId)
}

func (agentService *AgentService) watchMetaChange() {
	for meta := range agentService.failureChan {
		agentService.Lock()

		meta.ContainerName = strings.TrimPrefix(meta.ContainerName, "/")

		slog.Info("receive failure contaier", "containername", meta.ContainerName, "error", meta.ErrorMsg, "state", meta.State)

		c, ok := agentService.failureContainers[meta.ContainerName]

		if !ok {
			if !meta.IsDelete {
				agentService.failureContainers[meta.ContainerName] = &model.ContainerInfo{
					ContainerName: meta.ContainerName,
					State:         meta.State,
					ErrorMsg:      meta.ErrorMsg,
				}
			}
		} else {
			if meta.IsDelete {
				delete(agentService.failureContainers, meta.ContainerName)
			} else {
				c.State = meta.State
				c.ErrorMsg = meta.ErrorMsg
			}
		}

		agentService.Unlock()
	}
}

func (agentService *AgentService) sendHealthRequest(ctx context.Context) error {
	// 获取 Docker Engine 信息
	dockerEngineInfo, err := agentService.controller.DockerEngine(ctx)
	if err != nil {
		return err
	}

	//本地容器信息
	containers := []*model.ContainerInfo{}
	agentService.RLock()
	reportNames := make(map[string]string)
	for _, containerInfo := range agentService.containers {
		if fc, ok := agentService.failureContainers[containerInfo.ContainerName]; ok {
			containerInfo.State = fc.State
			containerInfo.ErrorMsg = fc.ErrorMsg
		}
		containers = append(containers, containerInfo)
		reportNames[containerInfo.ContainerName] = ""
	}

	for _, containerInfo := range agentService.failureContainers {
		if _, ok := reportNames[containerInfo.ContainerName]; !ok {
			slog.Info("report failure container", "containername", containerInfo.ContainerName, "state", containerInfo.State)
			containers = append(containers, containerInfo)
		}
	}

	agentService.RUnlock()

	payload := &model.HostHealthRequest{
		HostInfo:     model.GetHostInfo(agentService.config.APIConfig.Bind),
		DockerEngine: *dockerEngineInfo,
		Containers:   containers,
	}

	// for _, containerInfo := range containers {
	// 	slog.Info("report container", "containername", containerInfo.ContainerName, "state", containerInfo.State, "error", containerInfo.ErrorMsg)
	// }

	return reqclient.PostRequest(agentService.httpClient, fmt.Sprintf("%s/api/health", agentService.config.ServerConfig.Host), payload)
}

func (agentService *AgentService) sendConfigValuesRequest(configNames []string) (map[string][]byte, error) {
	configPair := map[string][]byte{}
	for _, configName := range configNames {
		data, err := reqclient.GetRequest(agentService.httpClient, fmt.Sprintf("%s/api/config/%s", agentService.config.ServerConfig.Host, configName))
		if err != nil {
			return nil, err
		}
		configPair[configName] = data
	}
	return configPair, nil
}
