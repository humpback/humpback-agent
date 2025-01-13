package model

import (
	"github.com/docker/docker/api/types"
	"humpback-agent/pkg/utils"
)

type ContainerPort struct {
	BindIP      string `json:"bindIP"`
	PrivatePort uint16 `json:"privatePort"`
	PublicPort  uint16 `json:"publicPort"`
	Type        string `json:"type"`
}

type ContainerIP struct {
	NetworkID  string `json:"networkID"`
	EndpointID string `json:"endpointID"`
	Gateway    string `json:"gateway"`
	IPAddress  string `json:"ipAddress"`
}

type ContainerInfo struct {
	ContainerId   string          `json:"containerId"`
	ContainerName string          `json:"containerName"`
	State         string          `json:"state"`
	Status        string          `json:"status"`
	Network       string          `json:"network"`
	Image         string          `json:"image"`
	Command       string          `json:"command"`
	Created       int64           `json:"created"`
	Ports         []ContainerPort `json:"ports"`
	IPAddr        []ContainerIP   `json:"ipAddr"`
}

func ParseContainerInfo(container types.Container) *ContainerInfo {

	ports := []ContainerPort{}
	for _, bindPortPair := range container.Ports {
		ports = append(ports, ContainerPort{
			BindIP:      bindPortPair.IP,
			PublicPort:  bindPortPair.PublicPort,
			PrivatePort: bindPortPair.PrivatePort,
			Type:        bindPortPair.Type,
		})
	}

	ipAddrs := []ContainerIP{}
	for _, networkVal := range container.NetworkSettings.Networks {
		ipAddrs = append(ipAddrs, ContainerIP{
			NetworkID:  networkVal.NetworkID,
			EndpointID: networkVal.EndpointID,
			Gateway:    networkVal.Gateway,
			IPAddress:  networkVal.IPAddress,
		})
	}

	return &ContainerInfo{
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
}
