package config

import (
	"os"
	"strconv"
	"common/models"

	"github.com/astaxie/beego"
)

var config *models.Config

// Init - Load config info
func Init() {
	endpoint := os.Getenv("DOCKER_ENDPOINT")
	if endpoint == "" {
		endpoint = beego.AppConfig.DefaultString("DOCKER_ENDPOINT", "unix:///var/run/docker.sock")
	}

	apiVersion := os.Getenv("DOCKER_API_VERSION")
	if apiVersion == "" {
		apiVersion = beego.AppConfig.DefaultString("DOCKER_API_VERSION", "v1.20")
	}

	dockerRegistryAdd := os.Getenv("DOCKER_REGISTRY_ADDRESS")
	if dockerRegistryAdd == "" {
		dockerRegistryAdd = beego.AppConfig.DefaultString("DOCKER_REGISTRY_ADDRESS", "docker.neg")
	}

	var enableBuildImg bool

	if tempEnableBuildImg := os.Getenv("ENABLE_BUILD_IMAGE"); tempEnableBuildImg != "" {
		if tempEnableBuildImg == "1" || tempEnableBuildImg == "true" {
			enableBuildImg = true
		}
	} else {
		enableBuildImg = beego.AppConfig.DefaultBool("ENABLE_BUILD_IMAGE", false)
	}

	clusterEnabled := false
	enabled := os.Getenv("DOCKER_CLUSTER_ENABLED")
	if enabled == "" {
		clusterEnabled = beego.AppConfig.DefaultBool("DOCKER_CLUSTER_ENABLED", false)
	} else {
		var err error
		if clusterEnabled, err = strconv.ParseBool(enabled); err != nil {
			clusterEnabled = false
		}
	}

	clusterURIs := os.Getenv("DOCKER_CLUSTER_URIS")
	if clusterURIs == "" {
		clusterURIs = beego.AppConfig.DefaultString("DOCKER_CLUSTER_URIS", "zk://127.0.0.1:2181")
	}

	clusterName := os.Getenv("DOCKER_CLUSTER_NAME")
	if clusterName == "" {
		clusterName = beego.AppConfig.DefaultString("DOCKER_CLUSTER_NAME", "humpback/center")
	}

	clusterHeartBeat := os.Getenv("DOCKER_CLUSTER_HEARTBEAT")
	if clusterHeartBeat == "" {
		clusterHeartBeat = beego.AppConfig.DefaultString("DOCKER_CLUSTER_HEARTBEAT", "10s")
	}

	clusterTTL := os.Getenv("DOCKER_CLUSTER_TTL")
	if clusterTTL == "" {
		clusterTTL = beego.AppConfig.DefaultString("DOCKER_CLUSTER_TTL", "35s")
	}

	clusterPortsRange := os.Getenv("DOCKER_CLUSTER_PORTS_RANGE")
	if clusterPortsRange == "" {
		clusterPortsRange = beego.AppConfig.DefaultString("DOCKER_CLUSTER_PORTS_RANGE", "0-0")
	}

	var logLevel int
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		logLevel, _ = strconv.Atoi(envLogLevel)
	} else {
		logLevel = beego.AppConfig.DefaultInt("LOG_LEVEL", 3)
	}

	config = &models.Config{
		DockerEndPoint:          endpoint,
		DockerAPIVersion:        apiVersion,
		DockerRegistryAddress:   dockerRegistryAdd,
		EnableBuildImage:        enableBuildImg,
		DockerClusterEnabled:    clusterEnabled,
		DockerClusterURIs:       clusterURIs,
		DockerClusterName:       clusterName,
		DockerClusterHeartBeat:  clusterHeartBeat,
		DockerClusterTTL:        clusterTTL,
		DockerClusterPortsRange: clusterPortsRange,
		LogLevel:                logLevel,
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
