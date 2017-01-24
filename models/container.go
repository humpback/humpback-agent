package models

import (
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	units "github.com/docker/go-units"
)

var (
	RestartPolicyType = map[string]interface{}{
		"no":         true,
		"always":     true,
		"on-failure": true,
	}

	ActionType = map[string]interface{}{
		"start":   true,
		"stop":    true,
		"restart": true,
		"kill":    true,
		"pause":   true,
		"unpause": true,
		"rename":  true,
		"upgrade": true,
	}
)

// Container - define container info struct
type Container struct {
	ID                string           `json:"Id"`
	Image             string           `json:"Image"`
	Command           string           `json:"Command"`
	Name              string           `json:"Name"`
	Ports             []PortBinding    `json:"Ports"`
	Volumes           []VolumesBinding `json:"Volumes"`
	DNS               []string         `json:"Dns"`
	Env               []string         `json:"Env"`
	HostName          string           `json:"HostName"`
	NetworkMode       string           `json:"NetworkMode"`
	Status            interface{}      `json:"Status,omitempty"`
	RestartPolicy     string           `json:"RestartPolicy,omitempty"`
	RestartRetryCount int              `json:"RestartRetryCount,omitempty"`
	ExtraHosts        []string         `json:"Extrahosts"`
	CPUShares         int64            `json:"CPUShares,omitempty"`
	Memory            int64            `json:"Memory,omitempty"`
	SHMSize           int64            `json:"SHMSize,omitempty"`
	Links             []string         `json:"Links"`
	Ulimits           []*units.Ulimit  `json:"Ulimits"`
}

// PortBinding - define container port binding info struct
type PortBinding struct {
	PrivatePort int    `json:"PrivatePort"`
	PublicPort  int    `json:"PublicPort"`
	Type        string `json:"Type"`
	IP          string `json:"Ip"`
}

// VolumesBinding - define container volumes binding info struct
type VolumesBinding struct {
	ContainerVolume string `json:"ContainerVolume"`
	HostVolume      string `json:"HostVolume"`
	BindMount       string `json:"BindMount,omitempty"`
}

// ContainerOperate - define container http request struct
type ContainerOperate struct {
	Action    string `json:"Action"`
	Container string `json:"Container"`
	ImageTag  string `json:"ImageTag"`
	NewName   string `json:"NewName"`
}

// ContainerLog - define container logs info struct
type ContainerLog struct {
	Stdout []string `json:"StdOut"`
	Stderr []string `json:"StdErr"`
}

// ContainerStatsFromDocker - define container stats info from dockerapi
type ContainerStatsFromDocker struct {
	Network struct {
		RxBytes float64 `json:"rx_bytes"`
		TxBytes float64 `json:"tx_bytes"`
	} `json:"network"`

	PreCPUStats struct {
		CPUUsage struct {
			TotalUsage float64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage float64 `json:"system_cpu_usage"`
	} `json:"precpu_stats"`

	CPUStats struct {
		CPUUsage struct {
			TotalUsage float64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage float64 `json:"system_cpu_usage"`
	} `json:"cpu_stats"`

	MemoryStats struct {
		Usage float64 `json:"usage"`
		Limit float64 `json:"limit"`
	} `json:"memory_stats"`
}

// ContainerStats - define container stats info struct
type ContainerStats struct {
	CPUUsage    string `json:"CPUUsage"`
	MemoryUsage int64  `json:"MemoryUsage"`
	MemoryLimit int64  `json:"MemoryLimit"`
	NetworkIn   int64  `json:"NetworkIn"`
	NetworkOut  int64  `json:"NetowrkOut"`
}

// Parse - parse container from original container info
func (container *Container) Parse(origContainer *types.ContainerJSON) {
	container.ID = origContainer.ID
	container.Name = strings.Replace(origContainer.Name, "/", "", 1)
	container.Image = origContainer.Config.Image
	container.Env = origContainer.Config.Env
	container.DNS = origContainer.HostConfig.DNS
	container.HostName = origContainer.Config.Hostname
	container.NetworkMode = origContainer.HostConfig.NetworkMode.NetworkName()
	container.Status = origContainer.State

	command := origContainer.Path + strings.Join(origContainer.Args, " ")
	container.Command = strings.TrimLeft(command, " ")

	for item := range origContainer.Config.ExposedPorts {
		containerPort, _ := strconv.Atoi(item.Port())
		tempPortbinding := PortBinding{
			PrivatePort: containerPort,
			PublicPort:  0,
			Type:        item.Proto(),
			IP:          "0.0.0.0",
		}
		hostportBind := origContainer.NetworkSettings.Ports[item]
		if hostportBind == nil {
			hostportBind = origContainer.HostConfig.PortBindings[item]
		}
		if hostportBind != nil && len(hostportBind) > 0 {
			hostPort, _ := strconv.Atoi(hostportBind[0].HostPort)
			tempPortbinding.PublicPort = hostPort
			if hostportBind[0].HostIP != "" {
				tempPortbinding.IP = hostportBind[0].HostIP
			}
		}

		container.Ports = append(container.Ports, tempPortbinding)
	}

	for _, mount := range origContainer.Mounts {
		tempVolumeBinding := VolumesBinding{
			ContainerVolume: mount.Destination,
			HostVolume:      mount.Source,
		}
		container.Volumes = append(container.Volumes, tempVolumeBinding)
	}

	container.RestartPolicy = origContainer.HostConfig.RestartPolicy.Name
	container.RestartRetryCount = origContainer.HostConfig.RestartPolicy.MaximumRetryCount
	container.ExtraHosts = origContainer.HostConfig.ExtraHosts
	container.CPUShares = origContainer.HostConfig.CPUShares
	container.Memory = origContainer.HostConfig.Memory / 1024 / 1024
	container.SHMSize = origContainer.HostConfig.ShmSize
	container.Links = origContainer.HostConfig.Links
	container.Ulimits = origContainer.HostConfig.Ulimits
}
