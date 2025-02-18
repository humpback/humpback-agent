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

type ContainerMeta struct {
	Image         string            `json:"image"`
	AlwaysPull    bool              `json:"alwaysPull"`
	Command       string            `json:"command"`
	Envs          []string          `json:"env"`
	Labels        map[string]string `json:"labels"`
	Network       *NetworkInfo      `json:"network"`
	RestartPolicy *RestartPolicy    `json:"restartPolicy"`
}
