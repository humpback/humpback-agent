package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"humpback-agent/config"
	"humpback-agent/controller"
	"humpback-agent/model"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type APIServer struct {
	svc *http.Server
	// tlsConfig *conf.TLSConfig
	shutdown bool
}

func NewAPIServer(controller controller.ControllerInterface, config *config.APIConfig, certBundle *model.CertificateBundle, token string, tokenChan chan string) (*APIServer, error) {

	router := NewRouter(controller, config, token, tokenChan)
	server := &APIServer{
		svc: &http.Server{
			Addr:         fmt.Sprintf(":%s", config.Port),
			Handler:      router,
			WriteTimeout: 90 * time.Second,
			ReadTimeout:  30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		shutdown: false,
	}

	if certBundle != nil {
		cert, _ := tls.X509KeyPair(certBundle.CertPEM, certBundle.KeyPEM)
		config := &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      certBundle.CertPool,
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    certBundle.CertPool,
		}
		server.svc.TLSConfig = config
	}
	return server, nil
}

func (server *APIServer) Startup(ctx context.Context) error {
	server.shutdown = false
	startCtx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	var err error
	errCh := make(chan error)
	go func(errCh chan<- error) {
		// if server.tlsConfig != nil {
		// 	logger.INFO("[API] server https TLS enabled.", server.svc.Addr)
		// 	e = server.svc.ListenAndServeTLS(server.tlsConfig.ServerCert, server.tlsConfig.ServerKey)
		// } else {
		e := server.svc.ListenAndServeTLS("", "")
		//}
		if !server.shutdown {
			errCh <- e
		}
	}(errCh)
	select {
	case <-startCtx.Done():
		logrus.Infof("[API] Server listening on [%s]...", server.svc.Addr)
	case err = <-errCh:
		if err != nil && !server.shutdown {
			logrus.Error("[API] Server listening failed.")
		}
	}
	close(errCh)
	return err
}

func (server *APIServer) Stop(ctx context.Context) error {
	logrus.Info("[API] Server stopping...")
	shutdownCtx, cancel := context.WithTimeout(ctx, time.Second*15)
	defer cancel()
	server.shutdown = true
	return server.svc.Shutdown(shutdownCtx)
}
