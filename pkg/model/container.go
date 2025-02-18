package model

import (
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"humpback-agent/pkg/utils"
	"strconv"
	"strings"
	"time"
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
	Ports         []ContainerPort `json:"ports"`
	IPAddr        []ContainerIP   `json:"ipAddr"`
	Created       int64           `json:"created"`
	Started       int64           `json:"started"`
	Finished      int64           `json:"finished"`
}

func ParseContainerInfo(container types.ContainerJSON) *ContainerInfo {
	createdTimestamp := int64(0)
	if createdAt, err := time.Parse(time.RFC3339Nano, container.Created); err == nil {
		createdTimestamp = createdAt.UnixMilli()
	}

	state, status := "", ""
	startedTimestamp, finishedTimestamp := int64(0), int64(0)
	if container.State != nil {
		state = container.State.Status
		startedAt, err := time.Parse(time.RFC3339Nano, container.State.StartedAt)
		if err == nil {
			startedTimestamp = startedAt.UnixMilli()
		}

		finishedAt, err := time.Parse(time.RFC3339Nano, container.State.FinishedAt)
		if err == nil {
			finishedTimestamp = finishedAt.UnixMilli()
		}

		if state != "exited" {
			status = utils.HumanDuration(time.Since(startedAt))
		} else {
			state = fmt.Sprintf("%s(%d)", state, container.State.ExitCode)
			status = utils.HumanDuration(time.Since(finishedAt))
		}
	}

	return &ContainerInfo{
		ContainerId:   container.ID,
		ContainerName: utils.ContainerName(container.Name),
		State:         state,
		Status:        status,
		Image:         container.Image,
		Network:       container.HostConfig.NetworkMode.NetworkName(),
		Command:       ParseContainerCommandWithConfig(container.Path, container.Config),
		Ports:         ParseContainerPortsWithPortMap(container.HostConfig.PortBindings),
		IPAddr:        ParseContainerIPAddrWithNetworkSettings(container.NetworkSettings),
		Created:       createdTimestamp,
		Started:       startedTimestamp,
		Finished:      finishedTimestamp,
	}
}

func ParseContainerCommandWithConfig(execPath string, containerConfig *container.Config) string {
	if containerConfig != nil {
		return fmt.Sprintf("%s %s", execPath, strings.Join(containerConfig.Cmd, " "))
	}
	return execPath
}

func ParseContainerPortsWithPortMap(portMap nat.PortMap) []ContainerPort {
	ports := []ContainerPort{}
	for containerVal, hostVal := range portMap {
		bindIP := ""
		privatePort := uint16(0)
		if len(hostVal) > 0 {
			bindIP = hostVal[0].HostIP
			if hp, err := strconv.Atoi(hostVal[0].HostPort); err == nil {
				privatePort = uint16(hp)
			}
		}

		publicPort := uint16(0)
		if hp, err := strconv.Atoi(containerVal.Port()); err == nil {
			publicPort = uint16(hp)
		}

		ports = append(ports, ContainerPort{
			BindIP:      bindIP,
			PublicPort:  publicPort,
			PrivatePort: privatePort,
			Type:        containerVal.Proto(),
		})
	}
	return ports
}

func ParseContainerIPAddrWithNetworkSettings(networkSettings *types.NetworkSettings) []ContainerIP {
	ipAddrs := []ContainerIP{}
	if networkSettings != nil {
		for _, network := range networkSettings.Networks {
			ipAddrs = append(ipAddrs, ContainerIP{
				NetworkID:  network.NetworkID,
				EndpointID: network.EndpointID,
				Gateway:    network.Gateway,
				IPAddress:  network.IPAddress,
			})
		}
	}
	return ipAddrs
}
