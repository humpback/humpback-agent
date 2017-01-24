package models

// Config - config info
type Config struct {
	AppVersion            string `json:"AppVersion"`
	DockerEndPoint        string `json:"DockerEndPoint`
	DockerAPIVersion      string `json:"DockerAPIVersion`
	DockerRegistryAddress string `json:"-"`
	EnableBuildImage      bool   `json:"-"`
	LogLevel              int    `json:"LogLevel"`
}
