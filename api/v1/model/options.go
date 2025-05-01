package model

type ContainerLogger struct {
	Driver  string            `json:"driver"`  // 日志驱动
	Options map[string]string `json:"options"` // 日志选项
}

// Network 配置
type ContainerNetwork struct {
	Mode        string            `json:"mode"`        // 网络模式（bridge, host, none）
	Hostname    string            `json:"hostname"`    // 主机名
	DomainName  string            `json:"domainName"`  // 域名
	MacAddress  string            `json:"macAddress"`  // MAC 地址
	IPv4Address string            `json:"ipv4Address"` // IPv4 地址
	IPv6Address string            `json:"ipv6Address"` // IPv6 地址
	DNS         []string          `json:"dns"`         // DNS 服务器
	Hosts       map[string]string `json:"hosts"`       // /etc/hosts 文件条目
}

// Runtime 配置
type ContainerRuntime struct {
	Privileged bool     `json:"privileged"` // 是否启用特权模式
	Init       bool     `json:"init"`       // 是否使用 init 进程
	Runtime    string   `json:"runtime"`    // 运行时（default, runc）
	Devices    []string `json:"devices"`    // 设备映射（hostDevice:containerDevice）
}

// Sysctls 配置
type ContainerSysctl map[string]string // 系统控制参数

// Resource Limits 配置
type ContainerResource struct {
	MemoryReserve int64 `json:"memoryReserve"` // 内存保留（字节）
	MemoryLimit   int64 `json:"memoryLimit"`   // 内存限制（字节）
	CPUQuota      int64 `json:"cpuQuota"`      // CPU 配额（微秒）
	Enable        bool  `json:"enable"`        // 是否启用 GPU
}

// Capabilities 配置
type ContainerCapability struct {
	Add  []string `json:"add"`  // 添加的 Capabilities
	Drop []string `json:"drop"` // 删除的 Capabilities
}
