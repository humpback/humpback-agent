package app

import (
	"context"
	"flag"
	"github.com/sirupsen/logrus"
	"humpback-agent/internal/api"
	"humpback-agent/internal/controller"
	"humpback-agent/internal/utils"
)

var (
	apiServer *api.APIServer
)

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

	dockerClient, err := buildDockerClient(appConfig.DockerConfig)
	if err != nil {
		logrus.Errorf("Build docker client error, %s", err.Error())
	}

	defer shutdown(ctx)
	appController := controller.NewController(dockerClient, appConfig.DockerTimeoutOpts.Request)
	apiServer, err = api.NewAPIServer(appController, appConfig.APIConfig)
	if err != nil {
		logrus.Errorf("Construct api server error, %s", err.Error())
		return
	}

	if err := apiServer.Startup(ctx); err != nil {
		logrus.Error("API server startup error, %s", err.Error())
		return
	}

	logrus.Info("Humpback Agent started.")
	utils.ProcessWaitForSignal(nil)
}

func shutdown(ctx context.Context) {
	if apiServer != nil {
		if err := apiServer.Stop(ctx); err != nil {
			logrus.Errorf("API server stop error, %s", err.Error())
		}
	}
	logrus.Info("Humpback Agent shutdown.")
}
