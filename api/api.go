package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"humpback-agent/config"
	"humpback-agent/interval/controller"

	"github.com/gin-gonic/gin"
)

type Router struct {
	engine     *gin.Engine
	httpSrv    *http.Server
	controller *controller.Controller
}

func InitRouter(ctl *controller.Controller) *Router {
	gin.SetMode(gin.ReleaseMode)
	r := &Router{engine: gin.New(), controller: ctl}
	r.setMiddleware()
	r.setRoute()
	return r
}

func (api *Router) Start(tlsFunc func() *tls.Config) {
	go func() {
		listeningAddress := fmt.Sprintf("0.0.0.0:%d", config.NodeArgs().Port)
		slog.Info("[Api] Listening...", "Address", listeningAddress)
		api.httpSrv = &http.Server{
			Addr:    listeningAddress,
			Handler: api.engine,
		}
		var err error
		if tlsFunc != nil {
			api.httpSrv.TLSConfig = &tls.Config{
				GetConfigForClient: func(info *tls.ClientHelloInfo) (*tls.Config, error) {
					return tlsFunc(), nil
				},
			}
			err = api.httpSrv.ListenAndServeTLS("", "")
		} else {
			err = api.httpSrv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			slog.Error("[Api] Listening failed", "Address", listeningAddress, "Error", err)
			os.Exit(98)
		}
	}()
}

func (api *Router) Close(c context.Context) error {
	return api.httpSrv.Shutdown(c)
}

func (api *Router) setMiddleware() {
	api.engine.Use(gin.Recovery(), log(), corsCheck())
}

func (api *Router) setRoute() {
	var routes = map[string]map[string][]any{
		"/api": {
			"/task": {handleRouteTask},
		},
	}

	for group, list := range routes {
		for path, fList := range list {
			routerGroup := api.engine.Group(fmt.Sprintf("%s%s", group, path), parseSliceAnyToSliceFunc(fList[:len(fList)-1])...)
			groupFunc := fList[len(fList)-1].(func(*gin.RouterGroup))
			groupFunc(routerGroup)
		}
	}
}

func parseSliceAnyToSliceFunc(functions []any) []gin.HandlerFunc {
	result := make([]gin.HandlerFunc, 0)
	for _, f := range functions {
		if fun, ok := f.(gin.HandlerFunc); ok {
			result = append(result, fun)
		}
	}
	return result
}
