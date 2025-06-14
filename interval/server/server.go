package server

import (
	"crypto/tls"
	"fmt"

	"humpback-agent/config"
	"humpback-agent/pkg/httpx"
	"humpback-agent/types"
)

type Server struct {
	httpx   httpx.HttpxClient
	tokenFn func() string
}

func NewServer(tlsFunc func() *tls.Config, tokenFn func() string) *Server {
	return &Server{
		httpx:   httpx.NewHttpXClient(tlsFunc, config.NodeArgs().RegisterToken),
		tokenFn: tokenFn,
	}
}

func (s *Server) Health(healthInfo *types.HealthInfo) (*types.HealthRespInfo, error) {
	result := &types.HealthRespInfo{}
	url := config.ParseServerAddress("/api/health")
	if err := s.httpx.Post(url, nil, s.newHeader(), healthInfo, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Server) Config(configNames []string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	for _, name := range configNames {
		var resp = make([]byte, 0)
		url := config.ParseServerAddress(fmt.Sprintf("/api/config/%s", name))
		if err := s.httpx.Get(url, nil, s.newHeader(), &resp); err != nil {
			return nil, err
		}
		result[name] = resp
	}
	return result, nil
}

func (s *Server) newHeader() map[string]string {
	header := map[string]string{
		"Content-Type": "application/json",
	}
	token := s.tokenFn()
	if token != "" {
		header["Authorization"] = fmt.Sprintf("Bearer %s", token)
	}
	return header
}
