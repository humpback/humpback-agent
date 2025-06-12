package config

import (
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"gopkg.in/yaml.v3"
)

var configuration *config

type config struct {
	Version string       `json:"version" yaml:"version"`
	Node    NodeConfig   `json:"node" yaml:"node"`
	Server  ServerConfig `json:"server" yaml:"server"`
	Docker  DockerConfig `json:"docker" yaml:"docker"`
}

type ServerConfig struct {
	Address string        `json:"address" yaml:"address" env:"SERVER_ADDRESS"`
	Timeout time.Duration `json:"timeout" yaml:"timeout" env:"SERVER_TIMEOUT"`
}

type DockerConfig struct {
	Host       string        `json:"host" yaml:"host" env:"DOCKER_HOST"`
	RootDir    string        `json:"rootDir" yaml:"rootDir" env:"DOCKER_ROOT_DIR"`
	ApiVersion string        `json:"apiVersion" yaml:"apiVersion" env:"DOCKER_API_VERSION"`
	Timeout    time.Duration `json:"timeout" yaml:"timeout" env:"DOCKER_TIMEOUT"`
}

type NodeConfig struct {
	HostIP         string        `json:"hostIp" yaml:"hostIp" env:"IP"`
	Port           int           `json:"port" yaml:"port" env:"PORT"`
	RegisterToken  string        `json:"registerToken" yaml:"registerToken" env:"REGISTER_TOKEN"`
	HealthInterval time.Duration `json:"healthInterval" yaml:"healthInterval" env:"HEALTH_INTERVAL"`
}

func InitConfig() error {
	if err := readConfigFile("./config/config.yaml"); err != nil {
		return err
	}
	if err := env.Parse(configuration); err != nil {
		return err
	}
	return nil
}

func readConfigFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read config file(%s), %s", filePath, err)
	}
	if err = yaml.Unmarshal(data, configuration); err != nil {
		return fmt.Errorf("config file(%s) unmarshal, %s", filePath, err)
	}
	return nil
}

func Config() any {
	return *configuration
}

func NodeArgs() NodeConfig {
	return configuration.Node
}

func ServerArgs() ServerConfig {
	return configuration.Server
}

func DockerArgs() DockerConfig {
	return configuration.Docker
}
