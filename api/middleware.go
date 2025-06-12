package api

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func log() gin.HandlerFunc {
	return func(c *gin.Context) {
		if (c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut) && strings.EqualFold(c.GetHeader("Content-Type"), "application/json") {
			data, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.Error(err)
				c.Abort()
				return
			}
			c.Set("Body", data)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(data))
		}
		startTime := time.Now()
		c.Next()
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			slog.Info("request", c.Request.Method, c.Request.URL, "T", time.Now().Sub(startTime).String())
			v, ok := c.Get("Body")
			if ok {
				fmt.Printf("%s\n", v)
			}
		}
	}
}

func corsCheck() gin.HandlerFunc {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	return cors.New(corsConfig)
}
