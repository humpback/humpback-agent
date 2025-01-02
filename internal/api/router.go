package api

import (
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	"humpback-agent/internal/api/factory"
	_ "humpback-agent/internal/api/v1/handler"
	"humpback-agent/internal/config"
	"humpback-agent/internal/controller"
)

type IRouter interface {
	//ServeHTTP used to handle the http requests
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type Router struct {
	engine   *gin.Engine
	handlers map[string]any
}

func engineMode(modeConfig string) string {
	mode := "debug"
	if modeConfig == "release" {
		mode = gin.ReleaseMode
	} else if modeConfig == "test" {
		mode = gin.TestMode
	} else {
		mode = gin.DebugMode
	}
	return mode
}

func NewRouter(controller controller.ControllerInterface, config *config.APIConfig) IRouter {
	gin.SetMode(engineMode(config.Mode))
	engine := gin.New()
	if gin.IsDebugging() {
		pprof.Register(engine)
	}
	middlewares := Middlewares(config.Middlewares...)
	engine.Use(middlewares...)
	engineHandlers := map[string]any{}
	for _, version := range config.Versions {
		routerHandler, err := factory.HandlerConstruct(version, controller)
		if err != nil {
			logrus.Errorf("[API] Router construct %s handler error, %s", version, err.Error())
			continue
		}
		engineHandlers[version] = routerHandler
	}

	router := &Router{
		engine:   engine,
		handlers: engineHandlers,
	}
	router.initRouter()
	return router
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.engine.ServeHTTP(w, r)
}

func (router *Router) initRouter() {
	for version, routerHandler := range router.handlers {
		if routerHandler != nil {
			routerHandler.(factory.Initializer).SetRouter(version, router.engine)
		}
	}
}
