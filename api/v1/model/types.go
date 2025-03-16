package model

type NetworkMode string

var (
	NetworkModeHost   NetworkMode = "host"
	NetworkModeBridge NetworkMode = "bridge"
	NetworkModeCustom NetworkMode = "custom"
)

type RestartPolicyMode string

var (
	RestartPolicyModeNo            RestartPolicyMode = "no"
	RestartPolicyModeAlways        RestartPolicyMode = "always"
	RestartPolicyModeOnFail        RestartPolicyMode = "on-failure"
	RestartPolicyModeUnlessStopped RestartPolicyMode = "unless-stopped"
)

type NetworkInfo struct {
	Mode        NetworkMode `json:"mode"`        // custom模式需要创建网络
	Hostname    string      `json:"hostname"`    // bridge及custom模式时可设置，用户容器的hostname
	NetworkName string      `json:"networkName"` //custom模式使用
	Ports       []*PortInfo `json:"ports"`
}

type PortInfo struct {
	HostPort      uint   `json:"hostPort"`
	ContainerPort uint   `json:"containerPort"`
	Protocol      string `json:"protocol"`
}

type RestartPolicy struct {
	Mode          RestartPolicyMode `json:"mode"`
	MaxRetryCount int               `json:"maxRetryCount"`
}

type ScheduleInfo struct {
	Timeout string   `json:"timeout"`
	Rules   []string `json:"rules"`
}

type Capabilities struct {
	CapAdd  *[]string `json:"capAdd"`
	CapDrop *[]string `json:"capDrop"`
}

type Resources struct {
	Memory            uint64 `json:"memory"`
	MemoryReservation uint64 `json:"memoryReservation"`
	MaxCpuUsage       uint64 `json:"maxCpuUsage"`
}

type LogConfig struct {
	Type   string            `json:"type"`
	Config map[string]string `json:"config"`
}

type ServiceVolumeType string

var (
	ServiceVolumeTypeBind   ServiceVolumeType = "bind"
	ServiceVolumeTypeVolume ServiceVolumeType = "volume"
)

type ServiceVolume struct {
	Type     ServiceVolumeType `json:"type"`
	Target   string            `json:"target"`
	Source   string            `json:"source"`
	Readonly bool              `json:"readOnly"`
}

type ContainerMeta struct {
	Image         string            `json:"image"`
	AlwaysPull    bool              `json:"alwaysPull"`
	Command       string            `json:"command"`
	Envs          []string          `json:"env"`
	Labels        map[string]string `json:"labels"`
	Volumes       []*ServiceVolume  `json:"volumes"`
	Network       *NetworkInfo      `json:"network"`
	RestartPolicy *RestartPolicy    `json:"restartPolicy"`
	Capabilities  *Capabilities     `json:"capabilities"`
	LogConfig     *LogConfig        `json:"logConfig"`
	Resources     *Resources        `json:"resources"`
	Privileged    bool              `json:"privileged"`
}
