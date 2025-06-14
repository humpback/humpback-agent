package node

import (
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"humpback-agent/config"
	"humpback-agent/types"
)

type Node struct{}

func NewNode() *Node {
	return &Node{}
}

func (n *Node) HostInfo() (*types.HostInfo, error) {
	cpuInfo, err := n.CpuInfo()
	if err != nil {
		return nil, fmt.Errorf("get cpu info failed: %s", err)
	}
	memInfo, err := n.MemoryInfo()
	if err != nil {
		return nil, fmt.Errorf("get memory info failed: %s", err)
	}
	host, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("get host info failed: %s", err)
	}
	ips, err := n.IPAddress()
	if err != nil {
		return nil, fmt.Errorf("get host ips failed: %s", err)
	}
	return &types.HostInfo{
		HostId:        host.HostID,
		Hostname:      host.Hostname,
		IpAddress:     ips,
		Port:          config.NodeArgs().Port,
		OSType:        host.OS,
		Platform:      fmt.Sprintf("%s %s", host.Platform, host.PlatformVersion),
		KernelVersion: host.KernelVersion,
		Cpu:           *cpuInfo,
		Memory:        *memInfo,
	}, nil
}

func (n *Node) CpuInfo() (*types.CpuInfo, error) {
	physicsCount, err := cpu.Counts(false)
	if err != nil {
		return nil, err
	}
	logicalCount, err := cpu.Counts(true)
	if err != nil {
		return nil, err
	}
	cpuInfo := &types.CpuInfo{
		PhysicsCount: physicsCount,
		LogicalCount: logicalCount,
	}
	percents, err := cpu.Percent(0, false)
	if err != nil {
		return nil, err
	}
	if len(percents) > 0 {
		cpuInfo.Percent = percents[0]
	}
	cpus, err := cpu.Info()
	if err != nil {
		return nil, err
	}
	for _, cpu := range cpus {
		cpuInfo.Names = append(cpuInfo.Names, cpu.ModelName)
	}

	return cpuInfo, nil
}

func (n *Node) MemoryInfo() (*types.MemoryInfo, error) {
	virtualMemoryStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	return &types.MemoryInfo{
		Total:   virtualMemoryStat.Total,
		Used:    virtualMemoryStat.Used,
		Percent: virtualMemoryStat.UsedPercent,
	}, nil
}

func (n *Node) IPAddress() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var ipAddresses []string
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		if isVirtualInterface(iface.Name) {
			continue
		}
		addresses, err := iface.Addrs()
		if err != nil {
			slog.Warn("Failed to get ip address.", "Interface", iface.Name, "Error", err)
			continue
		}

		for _, addr := range addresses {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
				ipAddresses = append(ipAddresses, ipNet.IP.String())
			}
		}
	}
	if len(ipAddresses) == 0 {
		return nil, fmt.Errorf("not found ip address")
	}
	return ipAddresses, nil
}

func isVirtualInterface(name string) bool {
	virtualPrefixes := []string{
		"docker",
		"veth",
		"br-",
		"cni",
		"flannel",
		"lo",
		"kube",
		"vmnet",
		"virbr",
		"zt",
		"tailscale",
		"vEthernet",
	}
	for _, prefix := range virtualPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}
