package app

import (
	"context"
	"flag"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"humpback-agent/config"
	"humpback-agent/pkg/utils"
	"humpback-agent/service"
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

func Bootstrap(ctx context.Context) {
	configFile := flag.String("f", "./config.yaml", "application configuration file path.")
	// 解析命令行参数
	flag.Parse()

	logrus.Info("Humpback Agent starting....")
	appConfig, err := loadConfig(*configFile)
	if err != nil {
		logrus.Errorf("Load application config error, %s", err.Error())
		return
	}

	if err := initLogger(appConfig.LoggerConfig); err != nil {
		logrus.Errorf("Init application logger error, %s", err.Error())
		return
	}

	agentService, err := service.NewAgentService(ctx, appConfig)
	if err != nil {
		logrus.Errorf("Init application agent service error, %s", err.Error())
		return
	}

	defer func() {
		agentService.Shutdown(ctx)
		logrus.Info("Humpback Agent shutdown.")
	}()

	logrus.Info("Humpback Agent started.")
	utils.ProcessWaitForSignal(nil)
}
