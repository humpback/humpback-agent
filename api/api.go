package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"humpback-agent/config"
	"humpback-agent/types"

	"github.com/gin-gonic/gin"
)

type Router struct {
	engine  *gin.Engine
	httpSrv *http.Server
}

func InitRouter() *Router {
	gin.SetMode(gin.ReleaseMode)
	r := &Router{engine: gin.New()}
	r.setMiddleware()
	r.setRoute()
	return r
}

func (api *Router) Start(certBundle *types.CertificateBundle) {
	go func() {
		listeningAddress := fmt.Sprintf("0.0.0.0:%d", config.NodeArgs().Port)
		slog.Info("[Api] Listening...", "Address", listeningAddress)
		api.httpSrv = &http.Server{
			Addr:    listeningAddress,
			Handler: api.engine,
		}
		var err error
		if certBundle != nil {
			cert, _ := tls.X509KeyPair(certBundle.CertPEM, certBundle.KeyPEM)
			api.httpSrv.TLSConfig = &tls.Config{
				Certificates: []tls.Certificate{cert},
				RootCAs:      certBundle.CertPool,
				ClientAuth:   tls.NoClientCert,
				ClientCAs:    certBundle.CertPool,
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
