package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"humpback-agent/api"
	v1model "humpback-agent/api/v1/model"
	"humpback-agent/config"
	"humpback-agent/controller"
	"humpback-agent/internal/docker"
	"humpback-agent/internal/schedule"
	"humpback-agent/model"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

type AgentService struct {
	sync.RWMutex
	config     *config.AppConfig
	apiServer  *api.APIServer
	httpClient *http.Client
	scheduler  schedule.TaskSchedulerInterface
	controller controller.ControllerInterface
	containers map[string]*model.ContainerInfo
}

func NewAgentService(ctx context.Context, config *config.AppConfig) (*AgentService, error) {
	//构建Docker-client
	dockerClient, err := docker.BuildDockerClient(config.DockerConfig)
	if err != nil {
		return nil, err
	}

	//构建API和Controller接口
	appController := controller.NewController(dockerClient, config.DockerTimeoutOpts.Request)
	apiServer, err := api.NewAPIServer(appController, config.APIConfig)
	if err != nil {
		return nil, err
	}

	//构建Agent服务
	agentService := &AgentService{
		config:    config,
		apiServer: apiServer,
		httpClient: &http.Client{
			Timeout: config.Health.Timeout,
		},
		scheduler:  schedule.NewJobScheduler(dockerClient), //构建任务定时调度器
		controller: appController,
		containers: make(map[string]*model.ContainerInfo),
	}

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
	agentService.Lock()
	for _, container := range containers {
		result = agentService.controller.Container().Get(ctx, &v1model.GetContainerRequest{ContainerId: container.ID})
		if result.Error != nil {
			return fmt.Errorf("load container inspect %s error, %v", container.ID, result.Error)
		}
		containerInfo := model.ParseContainerInfo(result.Object.(types.ContainerJSON))
		agentService.containers[container.ID] = containerInfo
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
					}
				}
				agentService.Lock()
				agentService.containers[containerInfo.ContainerId] = containerInfo
				agentService.Unlock()
				agentService.sendHealthRequest(context.Background())
			}
		case "destroy", "remove", "delete":
			if message.Action == "destroy" { //从job定时调度器删除, 无论是否在调度器中, 会自动处理
				agentService.removeFromScheduler(message.Actor.ID)
			}
			//修改容器状态
			agentService.Lock()
			if containerInfo, ret := agentService.containers[message.Actor.ID]; ret {
				containerInfo.State = model.ContainerStatusRemoved
			}
			agentService.Unlock()
			//主动通知一次心跳
			agentService.sendHealthRequest(context.Background())
			//缓存删除容器
			agentService.Lock()
			delete(agentService.containers, message.Actor.ID)
			agentService.Unlock()
		}
	}
}

func (agentService *AgentService) addToScheduler(containerId string, containerName string, containerImage string, containerLabels map[string]string) error {
	if value, ret := containerLabels[schedule.HumpbackJobRulesLabel]; ret {
		var (
			err        error
			timeout    time.Duration
			alwaysPull bool
		)
		rules := strings.Split(value, ";")
		if value, ret = containerLabels[schedule.HumpbackJobMaxTimeoutLabel]; ret {
			if timeout, err = time.ParseDuration(value); err != nil {
				return err
			}
		}

		if value, ret = containerLabels[schedule.HumpbackJobAlwaysPullLabel]; ret {
			if alwaysPull, err = strconv.ParseBool(value); err != nil {
				return err
			}
		}
		return agentService.scheduler.AddContainer(containerId, containerName, containerImage, alwaysPull, rules, timeout)
	}
	return nil
}

func (agentService *AgentService) removeFromScheduler(containerId string) error {
	return agentService.scheduler.RemoveContainer(containerId)
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
	for _, containerInfo := range agentService.containers {
		containers = append(containers, containerInfo)
	}
	agentService.RUnlock()

	hostHealthRequest := &model.HostHealthRequest{
		HostInfo:     model.GetHostInfo(agentService.config.APIConfig.Bind),
		DockerEngine: *dockerEngineInfo,
		Containers:   containers,
	}

	data, err := json.Marshal(hostHealthRequest)
	if err != nil {
		return err
	}

	// fmt.Println(string(data))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/health", agentService.config.ServerConfig.Host), bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := agentService.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("response status error %d", resp.StatusCode)
	}
	return nil
}
