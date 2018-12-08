package main

import "github.com/astaxie/beego"
import "github.com/astaxie/beego/plugins/cors"
import "github.com/humpback/common/models"
import "github.com/humpback/humpback-agent/config"
import "github.com/humpback/humpback-agent/controllers"
import "github.com/humpback/humpback-agent/routers"
import "github.com/humpback/humpback-center/cluster/types"

import (
	"os"
	"os/signal"
	"syscall"
)

func main() {

	config.Init()
	config.SetVersion("1.3.7")

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

	ipAddr, bindPort, err := config.GetNodeHTTPAddrIPPort()
	if err != nil {
		beego.Error("agent httpaddr error:" + err.Error())
	}

	beego.BConfig.Listen.HTTPPort = bindPort
	if conf.DockerClusterEnabled {
		clusterOptions := types.NewNodeRegisterOptions(ipAddr, bindPort, &conf)
		if err := types.NodeRegister(clusterOptions); err != nil {
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
		types.NodeClose()
		os.Exit(0)
	}
}
