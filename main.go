package main

import "github.com/astaxie/beego"
import "github.com/astaxie/beego/plugins/cors"
import "github.com/humpback/common/models"
import "github.com/humpback/humpback-agent/config"
import "github.com/humpback/humpback-agent/controllers"
import "github.com/humpback/humpback-agent/routers"
import "github.com/humpback/humpback-center/cluster"

import (
	"os"
	"os/signal"
	"syscall"
)

func main() {

	config.Init()
	config.SetVersion("1.3.3")

	controllers.Init()
	var conf = config.GetConfig()
	beego.BConfig.MaxMemory = conf.DockerComposePackageMaxSize
	composeStorage, err := models.NewComposeStorage(conf.DockerComposePath)
	if err != nil {
		beego.Error("compose storage error, " + err.Error())
		return
	}
	routers.Init(composeStorage)

	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Accept", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))
	beego.SetLogFuncCall(true)
	beego.SetLevel(conf.LogLevel)

	if conf.DockerClusterEnabled {
		clusterOptions := cluster.NewNodeRegisterOptions(beego.BConfig.Listen.HTTPPort,
			conf.DockerClusterName, conf.DockerClusterURIs, conf.DockerClusterHeartBeat,
			conf.DockerClusterTTL, conf.DockerAgentIPAddr, conf.DockerEndPoint, conf.DockerAPIVersion)
		if err := cluster.NodeRegister(clusterOptions); err != nil {
			beego.Error("cluster node register error:" + err.Error())
			return
		}
	}
	go signalListen()

	beego.Run()
}

func signalListen() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	for {
		<-c
		cluster.NodeClose()
		os.Exit(0)
	}
}
