package api

import (
	"humpback-agent/api/factory"
	"humpback-agent/config"
	"humpback-agent/controller"
	"log/slog"
	"net/http"
	"sync/atomic"

	"github.com/sirupsen/logrus"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	_ "humpback-agent/api/v1/handler"
)

type IRouter interface {
	//ServeHTTP used to handle the http requests
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type Router struct {
	engine   *gin.Engine
	handlers map[string]any
}

type TokenService struct {
	token atomic.Value
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

func tokenAuthMiddleware(c *gin.Context) {

	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		c.Abort()
		return
	}

	// 验证Bearer token格式
	if len(token) < 7 || token[:7] != "Bearer " {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		c.Abort()
		return
	}

	tokenString := token[7:]

	ts := c.MustGet("tokenService").(*TokenService)
	if ts.token.Load().(string) != tokenString {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		c.Abort()
		return
	}

	c.Next()
}

func NewRouter(controller controller.ControllerInterface, config *config.APIConfig, token string, tokenChan chan string) IRouter {
	gin.SetMode(engineMode(config.Mode))
	engine := gin.New()
	if gin.IsDebugging() {
		pprof.Register(engine)
	}
	middlewares := Middlewares(config.Middlewares...)
	engine.Use(middlewares...)

	tokenService := &TokenService{}
	tokenService.token.Store(token)

	go func() {
		for newToken := range tokenChan {
			slog.Info("[API] Update token")
			tokenService.token.Store(newToken)
		}
	}()

	engine.Use(func(c *gin.Context) {
		c.Set("tokenService", tokenService)
		c.Next()
	})
	engine.Use(tokenAuthMiddleware)

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
