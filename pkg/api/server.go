package api

import (
	"context"
	"github.com/sirupsen/logrus"
	"humpback-agent/pkg/config"
	"humpback-agent/pkg/controller"
	"net/http"
	"time"
)

type APIServer struct {
	svc *http.Server
	// tlsConfig *conf.TLSConfig
	shutdown bool
}

func NewAPIServer(controller controller.ControllerInterface, config *config.APIConfig) (*APIServer, error) {
	// var tlsConfig *tls.Config
	// if config.TLSConfig != nil {
	// 	certPool := x509.NewCertPool()
	// 	caCert, err := ioutil.ReadFile(config.TLSConfig.CaCert)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	certPool.AppendCertsFromPEM(caCert)
	// 	tlsConfig = &tls.Config{
	// 		ClientCAs:  certPool,
	// 		ClientAuth: tls.RequireAndVerifyClientCert,
	// 		NextProtos: []string{"http/1.1"},
	// 	}
	// }

	router := NewRouter(controller, config)
	server := &APIServer{
		svc: &http.Server{
			Addr:         config.Bind,
			Handler:      router,
			WriteTimeout: 90 * time.Second,
			ReadTimeout:  30 * time.Second,
			IdleTimeout:  60 * time.Second,
			// TLSConfig:    tlsConfig,
		},
		//tlsConfig: config.TLSConfig,
		shutdown: false,
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
		var e error
		// if server.tlsConfig != nil {
		// 	logger.INFO("[API] server https TLS enabled.", server.svc.Addr)
		// 	e = server.svc.ListenAndServeTLS(server.tlsConfig.ServerCert, server.tlsConfig.ServerKey)
		// } else {
		e = server.svc.ListenAndServe()
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
