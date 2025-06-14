package controller

import (
	"fmt"

	"humpback-agent/interval/docker"
	"humpback-agent/interval/node"
	"humpback-agent/interval/schedule"
	"humpback-agent/interval/security"
	"humpback-agent/interval/server"
)

type Controller struct {
	Node     *node.Node
	Docker   *docker.DockerDriver
	Schedule *schedule.Schedule
	Security *security.Security
	Server   *server.Server
	StopCh   chan struct{}
}

func NewController(stopCh chan struct{}) (*Controller, error) {
	docker, err := docker.NewDockerDriver()
	if err != nil {
		return nil, fmt.Errorf("New docker driver failed: %s", err)
	}
	security := security.NewSecurity()
	return &Controller{
		Node:     node.NewNode(),
		Docker:   docker,
		Schedule: schedule.NewSchedule(docker),
		Security: security,
		Server:   server.NewServer(security.GetClientTLS, security.GetToken),
		StopCh:   stopCh,
	}, nil
}

func (c *Controller) Start() error {
	if err := c.Security.StartupRegister(c.Node, c.StopCh); err != nil {
		return err
	}
	return nil
}
