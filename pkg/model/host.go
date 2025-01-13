package model

import (
	"fmt"
	"humpback-agent/pkg/utils"
	"os"
	"runtime"
)

type HostInfo struct {
	Hostname      string   `json:"hostName"`
	OSInformation string   `json:"osInformation"`
	KernelVersion string   `json:"kernelVersion"`
	TotalCPU      int      `json:"totalCPU"`
	UsedCPU       int      `json:"usedCPU"`
	CPUUsage      float32  `json:"cpuUsage"`
	TotalMemoryGB int      `json:"totalMemoryGB"`
	UsedMemoryGB  int      `json:"usedMemoryGB"`
	MemoryUsage   float32  `json:"memoryUsage"`
	HostIPs       []string `json:"hostIPs"`
	HostPort      int      `json:"hostPort"`
}

type DockerEngineInfo struct {
	Version        string   `json:"version"`
	APIVersion     string   `json:"apiVersion"`
	RootDirectory  string   `json:"rootDirectory"`
	StorageDriver  string   `json:"storageDriver"`
	LoggingDriver  string   `json:"loggingDriver"`
	VolumePlugins  []string `json:"volumePlugins"`
	NetworkPlugins []string `json:"networkPlugins"`
}

type HostHealthRequest struct {
	HostInfo     HostInfo         `json:"hostInfo"`
	DockerEngine DockerEngineInfo `json:"dockerEngine"`
	Containers   []*ContainerInfo `json:"containers"`
}

func GetHostInfo(bind string) HostInfo {
	hostname, _ := os.Hostname()
	osInfo := fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
	kernelVersion := utils.HostKernelVersion()
	totalCPU, usedCPU, cpuUsage := utils.HostCPU()
	totalMEM, usedMEM, memUsage := utils.HostMemory()
	return HostInfo{
		Hostname:      hostname,
		OSInformation: osInfo,
		KernelVersion: kernelVersion,
		TotalCPU:      totalCPU,
		UsedCPU:       usedCPU,
		CPUUsage:      cpuUsage,
		TotalMemoryGB: utils.BytesToGB(totalMEM),
		UsedMemoryGB:  utils.BytesToGB(usedMEM),
		MemoryUsage:   memUsage,
		HostIPs:       utils.HostIPs(),
		HostPort:      utils.BindPort(bind),
	}
}
