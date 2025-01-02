package app

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"humpback-agent/internal/config"
	"io"
	"os"
	"path/filepath"
)

func loadConfig(configPath string) (*config.AppConfig, error) {
	logrus.Info("Loading server config....")
	appConfig, err := config.NewAppConfig(configPath)
	if err != nil {
		return nil, err
	}

	logrus.Info("-----------------HUMPBACK AGENT CONFIG-----------------")
	logrus.Infof("API Bind: %s", appConfig.APIConfig.Bind)
	logrus.Infof("API Versions: %v", appConfig.APIConfig.Versions)
	logrus.Infof("API Middlewares: %v", appConfig.APIConfig.Middlewares)
	logrus.Infof("API Access Token: %s", appConfig.APIConfig.AccessToken)
	logrus.Infof("Docker Host: %s", appConfig.DockerConfig.Host)
	logrus.Infof("Docker Version: %s", appConfig.DockerConfig.Version)
	logrus.Infof("Docker AutoNegotiate: %v", appConfig.DockerConfig.AutoNegotiate)
	logrus.Info("-------------------------------------------------------")
	return appConfig, nil
}

func initLogger(loggerConfig *config.LoggerConfig) error {
	logDir := filepath.Dir(loggerConfig.LogFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	lumberjackLogger := &lumberjack.Logger{
		Filename:   loggerConfig.LogFile,
		MaxSize:    loggerConfig.MaxSize / 1024 / 1024,
		MaxBackups: loggerConfig.MaxBackups,
		MaxAge:     loggerConfig.MaxAge,
		Compress:   loggerConfig.Compress,
	}

	logrus.SetOutput(io.MultiWriter(os.Stdout, lumberjackLogger))
	level, err := logrus.ParseLevel(loggerConfig.Level)
	if err != nil {
		return err
	}

	logrus.SetLevel(level)
	if loggerConfig.Format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{})
	}
	return nil
}
