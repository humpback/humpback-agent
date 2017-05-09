package main

import (
	"humpback-agent/config"
	"humpback-agent/controllers"
	"humpback-agent/routers"
	"humpback-center/cluster"
	"os"
	"os/signal"
	"syscall"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/plugins/cors"
)

func main() {

	config.Init()
	controllers.Init()
	routers.Init()

<<<<<<< HEAD
	config.SetVersion("1.0.0")
=======
	config.SetVersion("1.1.2")
>>>>>>> a3577c9e4a312a6b1b7329b2e6f9e95fcb25a714

	var conf = config.GetConfig()
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
			conf.DockerClusterTTL, conf.DockerEndPoint, conf.DockerAPIVersion)
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
