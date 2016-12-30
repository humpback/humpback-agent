package main

import (
	"humpback-agent/config"
	"humpback-agent/controllers"
	"humpback-agent/routers"
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

	config.SetVersion("1.0.0")

	var conf = config.GetConfig()
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Accept", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))
	beego.SetLogFuncCall(true)
	beego.SetLevel(conf.LogLevel)

	go signalListen()

	beego.Run()
}

func signalListen() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	for {
		<-c
		os.Exit(0)
	}
}
