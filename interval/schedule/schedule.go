package schedule

import (
	"sync"

	"github.com/robfig/cron/v3"
	"humpback-agent/interval/docker"
)

type Schedule struct {
	l      sync.RWMutex
	c      *cron.Cron
	docker *docker.DockerDriver
	tasks  map[string]*Task
}

func NewSchedule(docker *docker.DockerDriver) *Schedule {
	return &Schedule{
		l:      sync.RWMutex{},
		c:      cron.New(),
		docker: docker,
		tasks:  make(map[string]*Task),
	}
}
