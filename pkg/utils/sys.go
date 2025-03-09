package utils

import (
	"fmt"
	"math"
	"net"
	"runtime"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"golang.org/x/sys/unix"
)

func HostCPU() (int, float32, float32) {
	totalCPU := runtime.NumCPU()
	// 获取 CPU 使用率
	percent, err := cpu.Percent(0, false)
	if err != nil {
		return totalCPU, 0.0, 0.0
	}
	// 计算 CPU 使用率
	cpuUsage := float32(math.Round(percent[0]*100) / 100) // CPU 使用率保留两位小数
	// 计算 UsedCPU 使用个数
	usedCPU := float32(totalCPU) * cpuUsage / 100
	return totalCPU, usedCPU, cpuUsage
}

func HostMemory() (uint64, float32, float32) {
	// 获取内存信息
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0.0, 0.0
	}

	totalMEM := memInfo.Total
	usedMEM := float32(memInfo.Used)
	memUsage := float32(math.Round(memInfo.UsedPercent*100) / 100) // 内存使用率保留两位小数
	return totalMEM, usedMEM, memUsage
}

func HostKernelVersion() string {
	var utsname unix.Utsname
	if err := unix.Uname(&utsname); err != nil {
		return "unknown"
	}

	n := 0
	for i, b := range utsname.Release {
		if b == 0 {
			break
		}
		n = i + 1
	}

	kernelVersion := ""
	if n > 0 {
		kernelVersion = string(utsname.Release[:n])
	} else {
		kernelVersion = string(utsname.Release[:])
	}
	return string(kernelVersion)
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

func ContainerName(name string) string {
	if strings.HasPrefix(name, "/") {
		return name[1:]
	}
	return name
}

func BytesToGB(size any) float32 {
	var ret float32
	switch v := size.(type) {
	case int:
		ret = float32(v) / 1024 / 1024 / 1024
	case int64:
		ret = float32(v) / 1024 / 1024 / 1024
	case uint64:
		ret = float32(v) / 1024 / 1024 / 1024
	case float32:
		ret = v / 1024 / 1024 / 1024
	case float64:
		ret = float32(v / 1024 / 1024 / 1024)
	}
	return float32(math.Round(float64(ret)*100) / 100)
}

func BindPort(bind string) int {
	if bind == "" {
		return 0
	}

	// 如果字符串包含 ":"，说明可能是 IP:Port 或 :Port 格式
	if strings.Contains(bind, ":") {
		// 按 ":" 分割字符串
		parts := strings.Split(bind, ":")
		// 取最后一个部分作为端口
		portStr := parts[len(parts)-1]
		// 将端口字符串转换为 int
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return 0
		}
		return port
	}

	// 如果字符串不包含 ":"，说明是纯端口号
	port, err := strconv.Atoi(bind)
	if err != nil {
		return 0
	}
	return port
}
