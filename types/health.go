package types

type HealthInfo struct {
	Host       HostInfo         `json:"host"`
	Docker     DockerInfo       `json:"docker"`
	Containers []*ContainerInfo `json:"containers"`
}

type HostInfo struct {
	HostId        string     `json:"hostId"`
	Hostname      string     `json:"hostName"`
	IpAddress     []string   `json:"ipAddress"`
	Port          uint64     `json:"port"`
	OSType        string     `json:"osType"`
	Platform      string     `json:"platform"`
	KernelVersion string     `json:"kernelVersion"`
	Cpu           CpuInfo    `json:"cpu"`
	Memory        MemoryInfo `json:"memory"`
}

type CpuInfo struct {
	PhysicsCount int      `json:"physicsCount"`
	LogicalCount int      `json:"logicalCount"`
	Percent      float64  `json:"percent"`
	Names        []string `json:"names"`
}

type MemoryInfo struct {
	Total   uint64  `json:"total"`
	Used    uint64  `json:"used"`
	Percent float64 `json:"percent"`
}

type DockerInfo struct {
	Id             string   `json:"id"`
	Name           string   `json:"name"`
	Version        string   `json:"version"`
	APIVersion     string   `json:"apiVersion"`
	MinAPIVersion  string   `json:"minAPIVersion"`
	DockerRootDir  string   `json:"dockerRootDir"`
	StorageDriver  string   `json:"storageDriver"`
	LoggingDriver  string   `json:"loggingDriver"`
	VolumePlugins  []string `json:"volumePlugins"`
	NetworkPlugins []string `json:"networkPlugins"`
	NCPU           int      `json:"ncpu"`
	MemTotal       int64    `json:"memTotal"`
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
	ErrorMsg      string            `json:"errorMsg"`
}

type MounteInfo struct {
	Source      string `json:"Source"`
	Destination string `json:"Destination"`
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

type HealthRespInfo struct {
	Token string `json:"token"`
}
