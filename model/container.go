package model

import (
	"fmt"
	"humpback-agent/pkg/utils"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
)

const (
	ContainerStatusPending  = "Pending"
	ContainerStatusStarting = "Starting"
	ContainerStatusCreated  = "Created"
	ContainerStatusRunning  = "Running"
	ContainerStatusFailed   = "Failed"
	ContainerStatusExited   = "Exited"
	ContainerStatusRemoved  = "Removed"
	ContainerStatusWarning  = "Warning"
)

var stateMap = map[string]string{
	"healthy":   ContainerStatusRunning,
	"unhealthy": ContainerStatusFailed,
	"starting":  ContainerStatusStarting,
	"running":   ContainerStatusRunning,
	"exited":    ContainerStatusExited,
	"create":    ContainerStatusCreated,
	"created":   ContainerStatusCreated,
	"stop":      ContainerStatusExited,
	"stopped":   ContainerStatusExited,
	"destroy":   ContainerStatusRemoved,
	"remove":    ContainerStatusRemoved,
	"delete":    ContainerStatusRemoved,
	"pending":   ContainerStatusPending,
	"warning":   ContainerStatusWarning,
}

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
		var err error
		var startedAt time.Time
		var finishedAt time.Time
		if container.State.Status != "created" {
			if startedAt, err = time.Parse(time.RFC3339Nano, container.State.StartedAt); err == nil {
				startedTimestamp = startedAt.UnixMilli()
			}

			if finishedAt, err = time.Parse(time.RFC3339Nano, container.State.FinishedAt); err == nil {
				finishedTimestamp = finishedAt.UnixMilli()
			}
		}

		if container.State.Status == "exited" {
			status = utils.HumanDuration(time.Since(finishedAt))
		} else {
			if container.State.Status != "created" {
				status = utils.HumanDuration(time.Since(startedAt))
			}
		}
	}

	state = stateMap[container.State.Status]
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

type DockerLog struct {
	Time   string `json:"time"`
	Stream string `json:"stream"`
	Log    string `json:"log"`
}

type DockerContainerLog struct {
	ContainerId string      `json:"containerId"`
	DockerLogs  []DockerLog `json:"containerLogs"`
}
