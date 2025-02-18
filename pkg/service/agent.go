package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"humpback-agent/pkg/api"
	v1model "humpback-agent/pkg/api/v1/model"
	"humpback-agent/pkg/config"
	"humpback-agent/pkg/controller"
	"humpback-agent/pkg/internal/docker"
	"humpback-agent/pkg/model"
	"net/http"
	"sync"
	"time"
)

type AgentService struct {
	sync.RWMutex
	config     *config.AppConfig
	apiServer  *api.APIServer
	httpClient *http.Client
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
	return agentService, nil
}

func (agentService *AgentService) Shutdown(ctx context.Context) {
	if agentService.apiServer != nil {
		if err := agentService.apiServer.Stop(ctx); err != nil {
			logrus.Errorf("Humpback Agent api server stop error, %s", err.Error())
		}
	}
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
				agentService.Lock()
				agentService.containers[containerInfo.ContainerId] = containerInfo
				agentService.Unlock()
				agentService.sendHealthRequest(context.Background())
			}
		case "destroy", "remove", "delete":
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
	if resp.StatusCode == http.StatusOK {
		return fmt.Errorf("response status error %d", resp.StatusCode)
	}
	return nil
}
