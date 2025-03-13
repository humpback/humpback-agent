package model

import (
	"fmt"
	"humpback-agent/pkg/utils"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
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
	"healthy":    ContainerStatusRunning,
	"unhealthy":  ContainerStatusFailed,
	"starting":   ContainerStatusStarting,
	"restarting": ContainerStatusStarting,
	"running":    ContainerStatusRunning,
	"exited":     ContainerStatusExited,
	"create":     ContainerStatusCreated,
	"created":    ContainerStatusCreated,
	"stop":       ContainerStatusExited,
	"stopped":    ContainerStatusExited,
	"destroy":    ContainerStatusRemoved,
	"remove":     ContainerStatusRemoved,
	"removing":   ContainerStatusRemoved,
	"delete":     ContainerStatusRemoved,
	"pending":    ContainerStatusPending,
	"warning":    ContainerStatusWarning,
}

type ContainerPort struct {
	BindIP      string `json:"bindIP"`
	PrivatePort int    `json:"privatePort"`
	PublicPort  int    `json:"publicPort"`
	Type        string `json:"type"`
}

type ContainerIP struct {
	NetworkID  string `json:"networkID"`
	EndpointID string `json:"endpointID"`
	Gateway    string `json:"gateway"`
	IPAddress  string `json:"ipAddress"`
}

type MounteInfo struct {
	Source      string `json:"Source"`
	Destination string `json:"Destination"`
}

type ContainerInfo struct {
	ContainerId   string            `json:"containerId"`
	ContainerName string            `json:"containerName"`
	State         string            `json:"state"`
	Status        string            `json:"status"`
	Network       string            `json:"network"`
	Image         string            `json:"image"`
	Labels        map[string]string `json:"labels"`
	Env           []string          `json:"env"`
	Mountes       []MounteInfo      `json:"mounts"`
	Command       string            `json:"command"`
	Ports         []ContainerPort   `json:"ports"`
	IPAddr        []ContainerIP     `json:"ipAddr"`
	Created       int64             `json:"created"`
	Started       int64             `json:"started"`
	Finished      int64             `json:"finished"`
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
	if state == "" {
		slog.Info("unknow container status", "status", container.State.Status)
	}
	return &ContainerInfo{
		ContainerId:   container.ID,
		ContainerName: utils.ContainerName(container.Name),
		State:         state,
		Status:        status,
		Image:         container.Config.Image,
		Labels:        container.Config.Labels,
		Network:       container.HostConfig.NetworkMode.NetworkName(),
		Env:           container.Config.Env,
		Mountes:       ParseContainerMountes(container.Mounts),
		Command:       ParseContainerCommandWithConfig(container.Path, container.Config),
		Ports:         ParseContainerPortsWithNetworkSettings(container.NetworkSettings),
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

func ParseContainerPortsWithNetworkSettings(networkSettings *types.NetworkSettings) []ContainerPort {
	ports := []ContainerPort{}
	for containerPort, bindings := range networkSettings.Ports {

		for _, binding := range bindings {
			//fmt.Printf("  Host IP: %s, Host Port: %s\n", binding.HostIP, binding.HostPort)

			portInfo := ContainerPort{
				BindIP: binding.HostIP,
				Type:   containerPort.Proto(),
			}

			pport, _ := strconv.Atoi(containerPort.Port())
			portInfo.PrivatePort = pport

			hport, _ := strconv.Atoi(binding.HostPort)
			portInfo.PublicPort = hport

			ports = append(ports, portInfo)
		}

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

func ParseContainerMountes(mps []types.MountPoint) []MounteInfo {
	if mps == nil {
		return nil
	}

	mountes := make([]MounteInfo, 0)
	for _, mp := range mps {
		m := MounteInfo{
			Source:      mp.Source,
			Destination: mp.Destination,
		}
		mountes = append(mountes, m)
	}
	return mountes
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

type ContainerNetwork struct {
	Name    string `json:"name"`
	RxBytes uint64 `json:"rxBytes"`
	TxBytes uint64 `json:"txBytes"`
}

type ContainerStats struct {
	CPUPercent       float64            `json:"cpuPercent"`
	MemoryUsageBytes uint64             `json:"memoryUsageBytes"`
	MemoryLimitBytes uint64             `json:"memoryLimitBytes"`
	DiskReadBytes    uint64             `json:"diskReadBytes"`
	DiskWriteBytes   uint64             `json:"diskWriteBytes"`
	Networks         []ContainerNetwork `json:"networks"`
	StatsTime        string             `json:"statsTime"`
}

func ParseContainerStats(containerStats *container.StatsResponse) *ContainerStats {
	stats := ContainerStats{}
	if containerStats != nil {
		stats.CPUPercent = calculateCPUPercent(containerStats)
		stats.MemoryUsageBytes = containerStats.MemoryStats.Usage
		stats.MemoryLimitBytes = containerStats.MemoryStats.Limit
		stats.Networks = calculateNetworkIO(containerStats)
		stats.DiskReadBytes, stats.DiskWriteBytes = calculateDiskIO(containerStats)
		stats.StatsTime = time.Now().Format(time.RFC3339Nano) //containerStats.Read.Format(time.RFC3339Nano)
	}
	return &stats
}

// 计算容器 CPU 使用率
func calculateCPUPercent(stats *container.StatsResponse) float64 {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage) - float64(stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage) - float64(stats.PreCPUStats.SystemUsage)
	// 获取 CPU 核心数
	cpuCount := float64(stats.CPUStats.OnlineCPUs)
	if cpuCount == 0 {
		//如果 OnlineCPUs 为 0，使用PercpuUsage的长度作为备选
		cpuCount = float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
		if cpuCount == 0 {
			// 如果仍然为 0，默认设置为 1（避免除零错误）
			cpuCount = 1.0
		}
	}
	// CPU 使用率 = (容器 CPU 使用时间增量 / 系统 CPU 时间增量) * CPU 核心数 * 100
	cpuPercent := (cpuDelta / systemDelta) * cpuCount * 100.0
	//cpuPercent := (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	return math.Round(cpuPercent*100) / 100 //保留两位小数
}

// 计算容器网络 I/O
func calculateNetworkIO(stats *container.StatsResponse) []ContainerNetwork {
	networks := []ContainerNetwork{}
	for name, network := range stats.Networks {
		networks = append(networks, ContainerNetwork{
			Name:    name,
			RxBytes: network.RxBytes,
			TxBytes: network.TxBytes,
		})
	}
	return networks
}

func calculateDiskIO(stats *container.StatsResponse) (uint64, uint64) {
	var diskReadBytes, diskWriteBytes uint64
	if len(stats.BlkioStats.IoServiceBytesRecursive) > 0 {
		for _, bytesIO := range stats.BlkioStats.IoServiceBytesRecursive {
			if bytesIO.Op == "read" {
				diskReadBytes = bytesIO.Value
			} else if bytesIO.Op == "write" {
				diskWriteBytes = bytesIO.Value
			}
		}
	}
	return diskReadBytes, diskWriteBytes
}
