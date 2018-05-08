package config

import "github.com/humpback/common/models"
import "github.com/astaxie/beego"

import (
	"os"
	"strconv"
)

var config *models.Config

// Init - Load config info
func Init() {
	envEndpoint := os.Getenv("DOCKER_ENDPOINT")
	if envEndpoint == "" {
		envEndpoint = beego.AppConfig.DefaultString("DOCKER_ENDPOINT", "unix:///var/run/docker.sock")
	}

	envAPIVersion := os.Getenv("DOCKER_API_VERSION")
	if envAPIVersion == "" {
		envAPIVersion = beego.AppConfig.DefaultString("DOCKER_API_VERSION", "v1.20")
	}

	envRegistryAddr := os.Getenv("DOCKER_REGISTRY_ADDRESS")
	if envRegistryAddr == "" {
		envRegistryAddr = beego.AppConfig.DefaultString("DOCKER_REGISTRY_ADDRESS", "docker.neg")
	}

	//若没有设置环境变量, 默认(0.0.0.0)时，则在节点注册时自动获取一个本机地址.
	envAgentIPAddr := os.Getenv("DOCKER_AGENT_IPADDR")
	if envAgentIPAddr == "" {
		envAgentIPAddr = beego.AppConfig.DefaultString("DOCKER_AGENT_IPADDR", "0.0.0.0")
	}

	var envEnableBuildImg bool
	if tempEnableBuildImg := os.Getenv("ENABLE_BUILD_IMAGE"); tempEnableBuildImg != "" {
		if tempEnableBuildImg == "1" || tempEnableBuildImg == "true" {
			envEnableBuildImg = true
		}
	} else {
		envEnableBuildImg = beego.AppConfig.DefaultBool("ENABLE_BUILD_IMAGE", false)
	}

	envComposePath := os.Getenv("DOCKER_COMPOSE_PATH")
	if envComposePath == "" {
		envComposePath = beego.AppConfig.DefaultString("DOCKER_COMPOSE_PATH", "./compose_files")
	}

	var envComposePackageMaxSize int64
	packageMaxSize := os.Getenv("DOCKER_COMPOSE_PACKAGE_MAXSIZE")
	if packageMaxSize == "" {
		envComposePackageMaxSize = beego.AppConfig.DefaultInt64("DOCKER_COMPOSE_PACKAGE_MAXSIZE", 67108864)
	} else {
		value, err := strconv.ParseInt(packageMaxSize, 10, 64)
		if err != nil {
			envComposePackageMaxSize = 67108864
		} else {
			envComposePackageMaxSize = value
		}
	}

	envClusterEnabled := false
	enabled := os.Getenv("DOCKER_CLUSTER_ENABLED")
	if enabled == "" {
		envClusterEnabled = beego.AppConfig.DefaultBool("DOCKER_CLUSTER_ENABLED", false)
	} else {
		var err error
		if envClusterEnabled, err = strconv.ParseBool(enabled); err != nil {
			envClusterEnabled = false
		}
	}

	envClusterURIs := os.Getenv("DOCKER_CLUSTER_URIS")
	if envClusterURIs == "" {
		envClusterURIs = beego.AppConfig.DefaultString("DOCKER_CLUSTER_URIS", "zk://127.0.0.1:2181")
	}

	envClusterName := os.Getenv("DOCKER_CLUSTER_NAME")
	if envClusterName == "" {
		envClusterName = beego.AppConfig.DefaultString("DOCKER_CLUSTER_NAME", "humpback/center")
	}

	envClusterHeartBeat := os.Getenv("DOCKER_CLUSTER_HEARTBEAT")
	if envClusterHeartBeat == "" {
		envClusterHeartBeat = beego.AppConfig.DefaultString("DOCKER_CLUSTER_HEARTBEAT", "10s")
	}

	envClusterTTL := os.Getenv("DOCKER_CLUSTER_TTL")
	if envClusterTTL == "" {
		envClusterTTL = beego.AppConfig.DefaultString("DOCKER_CLUSTER_TTL", "35s")
	}

	envClusterPortsRange := os.Getenv("DOCKER_CLUSTER_PORTS_RANGE")
	if envClusterPortsRange == "" {
		envClusterPortsRange = beego.AppConfig.DefaultString("DOCKER_CLUSTER_PORTS_RANGE", "0-0")
	}

	var logLevel int
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		logLevel, _ = strconv.Atoi(envLogLevel)
	} else {
		logLevel = beego.AppConfig.DefaultInt("LOG_LEVEL", 3)
	}

	config = &models.Config{
		DockerEndPoint:              envEndpoint,
		DockerAPIVersion:            envAPIVersion,
		DockerRegistryAddress:       envRegistryAddr,
		EnableBuildImage:            envEnableBuildImg,
		DockerComposePath:           envComposePath,
		DockerComposePackageMaxSize: envComposePackageMaxSize,
		DockerAgentIPAddr:           envAgentIPAddr,
		DockerClusterEnabled:        envClusterEnabled,
		DockerClusterURIs:           envClusterURIs,
		DockerClusterName:           envClusterName,
		DockerClusterHeartBeat:      envClusterHeartBeat,
		DockerClusterTTL:            envClusterTTL,
		DockerClusterPortsRange:     envClusterPortsRange,
		LogLevel:                    logLevel,
	}
}

// GetConfig - return config struct
func GetConfig() models.Config {
	return *config
}

// SetAppVersion - ser app version
func SetVersion(version string) {
	config.AppVersion = version
}
