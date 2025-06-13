package server

import (
	"humpback-agent/config"
)

type Server struct {
	httpx HttpxClient
}

func NewServer() *Server {
	return &Server{
		httpx: NewHttpXClient(nil, config.NodeArgs().RegisterToken),
	}
}

func RegisterNode() {
	httpC := NewHttpXClient(nil, config.NodeArgs().RegisterToken)

}

func ReportHealth() {

}

func GetConfigByName() {

}
