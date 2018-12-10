package config

import "github.com/astaxie/beego"
import "github.com/humpback/common/models"
import "github.com/humpback/gounits/network"

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

var config *models.Config
var enableAuthorization bool
var AuthorizationToken string

/**
Get config from env, if empty, get it from app config.
*/
func getConfigFromEnvOrAppConfig(key string, defaulValue string) string {
	tmpValue := os.Getenv(key)
	if tmpValue == "" {
		tmpValue = beego.AppConfig.DefaultString(key, defaulValue)
	}
	return tmpValue
}

/**
Init - Load config info
*/
func Init() {
	envEndpoint := getConfigFromEnvOrAppConfig("DOCKER_ENDPOINT", "unix:///var/run/docker.sock")
	envAPIVersion := getConfigFromEnvOrAppConfig("DOCKER_API_VERSION", "v1.20")
	envRegistryAddr := getConfigFromEnvOrAppConfig("DOCKER_REGISTRY_ADDRESS", "docker.neg")
	envNodeHTTPAddr := getConfigFromEnvOrAppConfig("DOCKER_NODE_HTTPADDR", "0.0.0.0:8500")
	envContainerPortsRange := getConfigFromEnvOrAppConfig("DOCKER_CONTAINER_PORTS_RANGE", "0-0")
	envComposePath := getConfigFromEnvOrAppConfig("DOCKER_COMPOSE_PATH", "./compose_files")
	AuthorizationToken = getConfigFromEnvOrAppConfig("AUTHORIZATION_TOKEN", "humpback")

	// enable authorization
	enableAuthorization = false
	if enableAuth := os.Getenv("ENABLE_AUTHORIZATION"); enableAuth != "" {
		if enableAuth == "1" || enableAuth == "true" {
			enableAuthorization = true
		}
	} else {
		enableAuthorization = beego.AppConfig.DefaultBool("ENABLE_AUTHORIZATION", false)
	}
	var envEnableBuildImg bool
	if tempEnableBuildImg := os.Getenv("ENABLE_BUILD_IMAGE"); tempEnableBuildImg != "" {
		if tempEnableBuildImg == "1" || tempEnableBuildImg == "true" {
			envEnableBuildImg = true
		}
	} else {
		envEnableBuildImg = beego.AppConfig.DefaultBool("ENABLE_BUILD_IMAGE", false)
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

	// Cluster config
	envClusterURIs := getConfigFromEnvOrAppConfig("DOCKER_CLUSTER_URIS", "zk://127.0.0.1:2181")
	envClusterName := getConfigFromEnvOrAppConfig("DOCKER_CLUSTER_NAME", "humpback/center")
	envClusterHeartBeat := getConfigFromEnvOrAppConfig("DOCKER_CLUSTER_HEARTBEAT", "10s")
	envClusterTTL := getConfigFromEnvOrAppConfig("DOCKER_CLUSTER_TTL", "35s")

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
		DockerNodeHTTPAddr:          envNodeHTTPAddr,
		DockerContainerPortsRange:   envContainerPortsRange,
		DockerClusterEnabled:        envClusterEnabled,
		DockerClusterURIs:           envClusterURIs,
		DockerClusterName:           envClusterName,
		DockerClusterHeartBeat:      envClusterHeartBeat,
		DockerClusterTTL:            envClusterTTL,
		LogLevel:                    logLevel,
	}
}

// GetConfig - return config struct
func GetConfig() models.Config {
	return *config
}

func GetEnableAuthorization() bool {
	return enableAuthorization
}

// SetVersion - set app version
func SetVersion(version string) {
	config.AppVersion = version
}

// GetNodeHTTPAddrIPPort - return local agent httpaddr info
func GetNodeHTTPAddrIPPort() (string, int, error) {

	httpAddr := strings.TrimSpace(config.DockerNodeHTTPAddr)
	if strings.Index(httpAddr, ":") < 0 {
		httpAddr = httpAddr + ":"
	}

	pAddrStr := strings.SplitN(httpAddr, ":", 2)
	if pAddrStr[1] == "" {
		httpAddr = httpAddr + "8500"
	}

	strHost, strPort, err := net.SplitHostPort(httpAddr)
	if err != nil {
		return "", 0, err
	}

	if strHost == "" || strHost == "0.0.0.0" {
		strHost = network.GetDefaultIP()
	}

	ip := net.ParseIP(strHost)
	if ip == nil {
		return "", 0, fmt.Errorf("httpAddr ip invalid")
	}

	port, err := strconv.Atoi(strPort)
	if err != nil || port > 65535 || port <= 0 {
		return "", 0, fmt.Errorf("httpAddr port invalid")
	}
	return ip.String(), port, nil
}
