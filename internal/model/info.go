package model

type HostInfo struct {
	Hostname      string   `json:"hostName"`
	OSInformation string   `json:"osInformation"`
	KernelVersion string   `json:"kernelVersion"`
	TotalCPU      int      `json:"totalCPU"`
	TotalMEM      uint64   `json:"totalMEM"`
	AvailableMEM  uint64   `json:"availableMEM"`
	FreeMEM       uint64   `json:"freeMEM"`
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

type ContainerPort struct {
	BindIP      string `json:"bindIP"`
	PrivatePort uint16 `json:"privatePort"`
	PublicPort  uint16 `json:"publicPort"`
	Type        string `json:"type"`
}

type ContainerIP struct {
	NetworkID  string `json:"networkID"`
	EndpointID string `json:"endpointID"`
	Gateway    string `json:"gateway"`
	IPAddress  string `json:"ipAddress"`
}

type ContainerInfo struct {
	ContainerId   string          `json:"containerId"`
	ContainerName string          `json:"containerName"`
	State         string          `json:"state"`
	Status        string          `json:"status"`
	Network       string          `json:"network"`
	Image         string          `json:"image"`
	Command       string          `json:"command"`
	Created       int64           `json:"created"`
	Ports         []ContainerPort `json:"ports"`
	IPAddr        []ContainerIP   `json:"ipAddr"`
}

type HostHealthRequest struct {
	HostInfo      HostInfo         `json:"hostInfo"`
	DockerEngine  DockerEngineInfo `json:"dockerEngine"`
	ContainerList []ContainerInfo  `json:"containerList"`
}
