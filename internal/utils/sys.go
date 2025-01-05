package utils

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/mem"
	"golang.org/x/sys/unix"
	"humpback-agent/internal/model"
	"net"
	"os"
	"runtime"
	"strings"
)

func HostMemory() (uint64, uint64, uint64) {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0, 0
	}
	return memInfo.Total, memInfo.Available, memInfo.Free
}

func HostKernelVersion() string {
	var utsname unix.Utsname
	if err := unix.Uname(&utsname); err != nil {
		return "unknown"
	}
	return string(utsname.Release[:])
}

func HostIPs() []string {
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Failed to get network interfaces:", err)
		return nil
	}

	var ipAddresses []string
	for _, iface := range interfaces {
		// 跳过回环接口和未启用的接口
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// 跳过 Docker 虚拟网卡
		if DockerInterface(iface.Name) {
			continue
		}

		addresses, err := iface.Addrs()
		if err != nil {
			fmt.Printf("Failed to get addresses for interface %s: %v\n", iface.Name, err)
			continue
		}

		for _, addr := range addresses {
			// 检查是否为 IPv4 地址
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					ipAddresses = append(ipAddresses, ipNet.IP.String())
				}
			}
		}
	}

	return ipAddresses
}

// 获取机器的真实有用 IP 地址（优先选择外部网络地址）
func HostIP() string {
	ipAddresses := HostIPs()
	if len(ipAddresses) == 0 {
		return ""
	}

	// 优先选择外部网络地址
	for _, ip := range ipAddresses {
		if isExternalIP(ip) {
			return ip
		}
	}
	// 如果没有外部网络地址，返回第一个地址
	return ipAddresses[0]
}

// 判断是否为 Docker 虚拟网卡
func DockerInterface(ifaceName string) bool {
	// Docker 虚拟网卡通常以以下前缀开头
	dockerPrefixes := []string{"docker", "veth", "br-", "cni", "flannel"}
	for _, prefix := range dockerPrefixes {
		if len(ifaceName) >= len(prefix) && ifaceName[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// 判断是否为外部网络地址
func isExternalIP(ip string) bool {
	// 外部网络地址通常为 192.168.x.x、10.x.x.x 或 172.16.x.x
	if len(ip) >= 7 && ip[:7] == "192.168" {
		return true
	}
	if len(ip) >= 3 && ip[:3] == "10." {
		return true
	}
	if len(ip) >= 4 && ip[:4] == "172." {
		return true
	}
	return false
}

func HostInfo() model.HostInfo {
	hostname, _ := os.Hostname()
	osInfo := fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
	kernelVersion := HostKernelVersion()
	totalCPU := runtime.NumCPU()
	totalMem, availableMem, freeMem := HostMemory()
	return model.HostInfo{
		Hostname:      hostname,
		OSInformation: osInfo,
		KernelVersion: kernelVersion,
		TotalCPU:      totalCPU,
		TotalMEM:      totalMem,
		AvailableMEM:  availableMem,
		FreeMEM:       freeMem,
		HostIPs:       HostIPs(),
	}
}

func ContainerName(names []string) string {
	if names == nil || len(names) == 0 {
		return ""
	}

	if strings.HasPrefix(names[0], "/") {
		return names[0][1:]
	}
	return names[0]
}
