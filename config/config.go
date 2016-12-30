package config

import (
	"humpback-agent/models"
	"os"
	"strconv"

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

	var logLevel int
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		logLevel, _ = strconv.Atoi(envLogLevel)
	} else {
		logLevel = beego.AppConfig.DefaultInt("LOG_LEVEL", 3)
	}

	config = &models.Config{
		DockerEndPoint:   endpoint,
		DockerAPIVersion: apiVersion,
		LogLevel:         logLevel,
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
