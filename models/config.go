package models

// Config - config info
type Config struct {
	AppVersion       string `json:"AppVersion"`
	DockerEndPoint   string `json:"DockerEndPoint`
	DockerAPIVersion string `json:"DockerAPIVersion`
	LogLevel         int    `json:"LogLevel"`
}
