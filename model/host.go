package model

import (
	"crypto/ecdsa"
	"crypto/x509"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"humpback-agent/pkg/utils"
)

type HostInfo struct {
	Hostname      string   `json:"hostName"`
	OSInformation string   `json:"osInformation"`
	KernelVersion string   `json:"kernelVersion"`
	TotalCPU      int      `json:"totalCPU"`
	UsedCPU       float32  `json:"usedCPU"`
	CPUUsage      float32  `json:"cpuUsage"`
	TotalMemory   uint64   `json:"totalMemory"`
	UsedMemory    uint64   `json:"usedMemory"`
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

type CertificateBundle struct {
	Cert     *x509.Certificate
	PrivKey  *ecdsa.PrivateKey
	CertPool *x509.CertPool // CA证书池
	CertPEM  []byte         // PEM编码的证书
	KeyPEM   []byte         // PEM编码的私钥
}

func GetHostInfo(hostIpStr, portStr string) HostInfo {
	hostname, _ := os.Hostname()
	port, _ := strconv.Atoi(portStr)
	var hostIps []string
	if hostIpStr != "" {
		hostIps = []string{hostIpStr}
	} else {
		hostIps = utils.HostIPs()
	}
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
		TotalMemory:   totalMEM,
		UsedMemory:    usedMEM,
		MemoryUsage:   memUsage,
		HostIPs:       hostIps,
		HostPort:      port,
	}
}
