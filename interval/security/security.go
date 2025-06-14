package security

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"humpback-agent/config"
	"humpback-agent/interval/node"
	"humpback-agent/pkg/httpx"
)

type registerResp struct {
	CertPEM  string `json:"certPem"`
	KeyPEM   string `json:"keyPem"`
	CAPEM    string `json:"caPem"`
	Token    string `json:"token"`
	ExpireAt int64  `json:"expireAt"`
}

type Security struct {
	l         sync.RWMutex
	httpx     httpx.HttpxClient
	token     atomic.Value
	serverTls atomic.Value
	clientTls atomic.Value
	ExpireAt  int64
}

func NewSecurity() *Security {
	return &Security{
		httpx: httpx.NewHttpXClient(nil, config.NodeArgs().RegisterToken),
	}
}

func (s *Security) StartupRegister(node *node.Node, stopCh chan struct{}) error {
	if err := s.register(node); err != nil {
		return fmt.Errorf("[Security] register node failed: %s", err)
	}
	go s.checkCertLoop(node, stopCh)
	return nil
}

func (s *Security) checkCertLoop(node *node.Node, stopCh chan struct{}) {
	duration := 24 * time.Hour
	slog.Info("[Security] start interval check cert.", "Duration", duration.String())
	tiker := time.NewTicker(duration)
	defer tiker.Stop()
	for {
		select {
		case <-tiker.C:
			if s.ExpireAt-time.Now().Unix() < int64((120 * time.Hour).Seconds()) {
				if err := s.register(node); err != nil {
					slog.Error("[Security] interval register node failed.", "Error", err)
					tiker.Reset(time.Hour)
				} else {
					tiker.Reset(duration)
				}
			}
		case <-stopCh:
			return
		}
	}
}

func (s *Security) register(node *node.Node) error {
	ips, err := node.IPAddress()
	if err != nil {
		return err
	}
	var (
		body = map[string]any{
			"hostIPs": ips,
			"token":   config.NodeArgs().RegisterToken,
		}
		data = &registerResp{}
	)
	url := config.ParseServerAddress("/api/register")
	if err = s.httpx.Post(url, nil, map[string]string{"Content-Type": "application/json"}, body, data); err != nil {
		return err
	}
	return s.parseCert(data)
}

func (s *Security) parseCert(regResp *registerResp) error {
	caBlock, _ := pem.Decode([]byte(regResp.CAPEM))
	if caBlock == nil {
		return errors.New("invalid CA certificate format")
	}
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %w", err)
	}
	certPool := x509.NewCertPool()
	certPool.AddCert(caCert)
	cert, err := tls.X509KeyPair([]byte(regResp.CertPEM), []byte(regResp.KeyPEM))
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}
	if regResp.Token != "" {
		s.token.Store(regResp.Token)
	}
	s.ExpireAt = regResp.ExpireAt
	s.serverTls.Store(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	})
	s.clientTls.Store(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
		ClientAuth:   tls.NoClientCert,
		ClientCAs:    certPool,
	})
	return nil
}

func (s *Security) GetServerTLS() *tls.Config {
	return s.serverTls.Load().(*tls.Config)
}

func (s *Security) GetClientTLS() *tls.Config {
	return s.serverTls.Load().(*tls.Config)
}

func (s *Security) GetToken() string {
	return s.token.Load().(string)
}

func (s *Security) SetToken(token string) {
	s.token.Store(token)
}
