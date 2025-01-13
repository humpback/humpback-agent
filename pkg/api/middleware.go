package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"strings"
)

var defaultMiddlewares = map[string]gin.HandlerFunc{
	"cors": cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"*"},
		AllowHeaders: []string{"*"},
	}),
	"recovery": gin.Recovery(),
	"logger":   gin.Logger(),
}

func MiddlewareCors() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"*"},
		AllowHeaders: []string{"*"},
	})
}

func MiddlewareLogger() gin.HandlerFunc {
	return gin.Logger()
}

func MiddlewareRecovery() gin.HandlerFunc {
	return gin.Recovery()
}

func Middleware(middleware string) gin.HandlerFunc {
	middleFunc, ret := defaultMiddlewares[middleware]
	if !ret {
		return nil
	}
	return middleFunc
}

func Middlewares(middleware ...string) []gin.HandlerFunc {
	middlewares := []gin.HandlerFunc{}
	for _, name := range middleware {
		name = strings.TrimSpace(strings.ToLower(name))
		if middleFunc, ret := defaultMiddlewares[name]; ret {
			middlewares = append(middlewares, middleFunc)
		}
	}
	return middlewares
}

func AllMiddlewares() []gin.HandlerFunc {
	middlewares := []gin.HandlerFunc{}
	for _, middlewareFunc := range defaultMiddlewares {
		middlewares = append(middlewares, middlewareFunc)
	}
	return middlewares
}
