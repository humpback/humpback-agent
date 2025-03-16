package config

import (
	"errors"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

var (
	ErrAPIConfigInvalid    = errors.New("api config invalid")
	ErrDockerConfigInvalid = errors.New("docker config invalid")

	defaultLogConfig = &LoggerConfig{
		LogFile:    "",
		Level:      "info",
		Format:     "json",
		MaxSize:    20971520,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   false,
	}

	defaultVolumesPath = "/var/lib/humpback/volumes"
)

type APIConfig struct {
	Bind        string   `json:"bind" yaml:"bind"`
	Mode        string   `json:"mode" yaml:"mode"`
	Middlewares []string `json:"middlewares" yaml:"middlewares"`
	Versions    []string `json:"versions" yaml:"versions"`
	AccessToken string   `json:"accessToken" yaml:"accessToken"`
}

type LoggerConfig struct {
	LogFile    string `json:"logFile" yaml:"logFile"`
	Level      string `json:"level" yaml:"level"`
	Format     string `json:"format" yaml:"format"`
	MaxSize    int    `json:"maxSize" yaml:"maxSize"`
	MaxAge     int    `json:"maxAge" yaml:"maxAge"`
	MaxBackups int    `json:"maxBackups" yaml:"maxBackups"`
	Compress   bool   `json:"compress" yaml:"compress"`
}

type DockerTimeoutOpts struct {
	Connection time.Duration `json:"connection" yaml:"connection"`
	Request    time.Duration `json:"request" yaml:"request"`
}

type DockerTLSOpts struct {
	Enabled            bool   `json:"enabled" yaml:"enabled"`
	CAPath             string `json:"caPath" yaml:"caPath"`
	CertPath           string `json:"certPath" yaml:"certPath"`
	KeyPath            string `json:"keyPath" yaml:"keyPath"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify" yaml:"insecureSkipVerify"`
}

type DockerRegistryOpts struct {
	Default  string `json:"default" yaml:"default"`
	UserName string `json:"userName" yaml:"userName"`
	Password string `json:"password" yaml:"password"`
}

type DockerConfig struct {
	Host               string             `json:"host" yaml:"host"`
	Version            string             `json:"version" yaml:"version"`
	AutoNegotiate      bool               `json:"autoNegotiate" yaml:"autoNegotiate"`
	DockerTimeoutOpts  DockerTimeoutOpts  `json:"timeout" yaml:"timeout"`
	DockerTLSOpts      DockerTLSOpts      `json:"tls" yaml:"tls"`
	DockerRegistryOpts DockerRegistryOpts `json:"registry" yaml:"registry"`
}

type ServerHealthConfig struct {
	Interval time.Duration `json:"interval" yaml:"interval"`
	Timeout  time.Duration `json:"timeout" yaml:"timeout"`
}

type ServerConfig struct {
	Host   string             `json:"host" yaml:"host"`
	Health ServerHealthConfig `json:"health" yaml:"health"`
}

type VolumesConfig struct {
	RootDirectory string `yaml:"rootDirectory"`
}

type AppConfig struct {
	*APIConfig     `json:"api" yaml:"api"`
	*ServerConfig  `json:"server" yaml:"server"`
	*VolumesConfig `json:"volumes" yaml:"volumes"`
	*DockerConfig  `json:"docker" yaml:"docker"`
	*LoggerConfig  `json:"logger" yaml:"logger"`
}

func NewAppConfig(configPath string) (*AppConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	appConfig := AppConfig{}
	if err = yaml.Unmarshal(data, &appConfig); err != nil {
		return nil, err
	}

	if err = ParseConfigFromEnv(&appConfig); err != nil {
		return nil, err
	}

	if appConfig.APIConfig == nil {
		return nil, ErrAPIConfigInvalid
	}

	if appConfig.VolumesConfig == nil || appConfig.VolumesConfig.RootDirectory == "" {
		appConfig.VolumesConfig = &VolumesConfig{
			RootDirectory: defaultVolumesPath,
		}
	}

	if appConfig.DockerConfig == nil {
		return nil, ErrDockerConfigInvalid
	}

	if appConfig.LoggerConfig == nil {
		appConfig.LoggerConfig = defaultLogConfig
	}
	return &appConfig, nil
}
