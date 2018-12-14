package models

// Config - config info
type Config struct {
	AppVersion                  string `json:"AppVersion"`
	DockerEndPoint              string `json:"DockerEndPoint"`
	DockerAPIVersion            string `json:"DockerAPIVersion"`
	DockerRegistryAddress       string `json:"-"`
	EnableBuildImage            bool   `json:"-"`
	DockerComposePath           string `json:"DockerComposePath"`
	DockerComposePackageMaxSize int64  `json:"DockerComposePackageMaxSize"`
	DockerNodeHTTPAddr          string `json:"DockerNodeHTTPAddr"`
	DockerContainerPortsRange   string `json:"DockerContainerPortsRange"`
	DockerClusterEnabled        bool   `json:"DockerClusterEnabled"`
	DockerClusterURIs           string `json:"DockerClusterURIs"`
	DockerClusterName           string `json:"DockerClusterName"`
	DockerClusterHeartBeat      string `json:"DockerClusterHeartBeat"`
	DockerClusterTTL            string `json:"DockerClusterTTL"`
	LogLevel                    int    `json:"LogLevel"`
}
