package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/sirupsen/logrus"
	v1model "humpback-agent/internal/api/v1/model"
	"humpback-agent/internal/config"
	"humpback-agent/internal/controller"
	"humpback-agent/internal/model"
	"humpback-agent/internal/utils"
	"net/http"
	"time"
)

type HealthClient struct {
	controller   controller.ControllerInterface
	serverConfig *config.ServerConfig
	httpClient   *http.Client
}

func NewHealthClient(controller controller.ControllerInterface, serverConfig *config.ServerConfig) *HealthClient {
	return &HealthClient{
		controller:   controller,
		serverConfig: serverConfig,
		httpClient: &http.Client{
			Timeout: serverConfig.Health.Timeout,
		},
	}
}

func (client *HealthClient) Heartbeat() {
	if err := client.sendHealthRequest(context.Background()); err != nil {
		logrus.Errorf("send health request error: %+v", err.Error())
	}
	go client.doExecLoop()
}

func (client *HealthClient) readHostInfo(ctx context.Context) (*model.HostHealthRequest, error) {
	// 获取机器基本信息
	hostInfo := utils.HostInfo()

	// 获取 Docker Engine 信息
	dockerEngineInfo, err := client.controller.DockerEngine(ctx)
	if err != nil {
		return nil, err
	}

	//获取容器列表
	result := client.controller.Container().List(ctx, &v1model.QueryContainerRequest{All: true})
	if result.Error != nil {
		return nil, fmt.Errorf("get containers failed, %s", result.Error.ErrMsg)
	}

	var containers []model.ContainerInfo
	for _, container := range result.Object.([]types.Container) {
		ports := []model.ContainerPort{}
		for _, bindPortPair := range container.Ports {
			ports = append(ports, model.ContainerPort{
				BindIP:      bindPortPair.IP,
				PublicPort:  bindPortPair.PublicPort,
				PrivatePort: bindPortPair.PrivatePort,
				Type:        bindPortPair.Type,
			})
		}

		ipAddrs := []model.ContainerIP{}
		for _, networkVal := range container.NetworkSettings.Networks {
			ipAddrs = append(ipAddrs, model.ContainerIP{
				NetworkID:  networkVal.NetworkID,
				EndpointID: networkVal.EndpointID,
				Gateway:    networkVal.Gateway,
				IPAddress:  networkVal.IPAddress,
			})
		}

		containerInfo := model.ContainerInfo{
			ContainerId:   container.ID,
			ContainerName: utils.ContainerName(container.Names),
			State:         container.State,
			Status:        container.Status,
			Image:         container.Image,
			Network:       container.HostConfig.NetworkMode,
			Command:       container.Command,
			Created:       container.Created,
			Ports:         ports,
			IPAddr:        ipAddrs,
		}
		containers = append(containers, containerInfo)
	}
	return &model.HostHealthRequest{
		HostInfo:      hostInfo,
		DockerEngine:  *dockerEngineInfo,
		ContainerList: containers,
	}, nil
}

func (client *HealthClient) sendHealthRequest(ctx context.Context) error {
	hostInfo, err := client.readHostInfo(ctx)
	if err != nil {
		return err
	}

	data, err := json.Marshal(hostInfo)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/health", client.serverConfig.Host), bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return fmt.Errorf("response status error %d", resp.StatusCode)
	}
	return nil
}

func (client *HealthClient) doExecLoop() {
	for {
		time.Sleep(client.serverConfig.Health.Interval)
		if err := client.sendHealthRequest(context.Background()); err != nil {
			logrus.Errorf("send health request error: %+v", err.Error())
			continue
		}
		logrus.Debugf("send health request done at %s\n", time.Now().String())
	}
}
