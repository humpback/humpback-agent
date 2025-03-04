package schedule

import (
	"fmt"
	"sync"
	"time"

	"github.com/docker/docker/client"
	"github.com/robfig/cron/v3"
)

const (
	HumpbackJobRulesLabel      = "HUMPBACK_JOB_RULES"
	HumpbackJobAlwaysPullLabel = "HUMPBACK_JOB_ALWAYS_PULL"
	HumpbackJobMaxTimeoutLabel = "HUMPBACK_JOB_MAX_TIMEOUT"
)

const (
	MaxKillContainerTimeoutSeconds = 5
)

type TaskSchedulerInterface interface {
	Start()
	Stop()
	AddContainer(containerId string, name string, image string, alwaysPull bool, rules []string, timeout time.Duration) error
	RemoveContainer(containerId string) error
}

type TaskScheduler struct {
	sync.RWMutex
	c      *cron.Cron
	client *client.Client
	tasks  map[cron.EntryID]*Task //entryId, *task
}

func NewJobScheduler(client *client.Client) TaskSchedulerInterface {
	return &TaskScheduler{
		c:      cron.New(),
		client: client,
		tasks:  make(map[cron.EntryID]*Task),
	}
}

func (scheduler *TaskScheduler) Start() {
	scheduler.c.Start()
}

func (scheduler *TaskScheduler) Stop() {
	scheduler.c.Stop()
}

func (scheduler *TaskScheduler) AddContainer(containerId string, name string, image string, alwaysPull bool, rules []string, timeout time.Duration) error {
	scheduler.Lock()
	defer scheduler.Unlock()
	//同名的容器不能反复进入调度器, 因为可能是dockerEvent捕获到了task内部因AlwaysPull导致的容器替换reCreate
	for _, task := range scheduler.tasks {
		if task.Name == name {
			return fmt.Errorf("container %s already exists in scheduler", containerId)
		}
	}

	for _, rule := range rules {
		task := NewTask(containerId, name, image, alwaysPull, timeout, rule, scheduler.client)
		entryId, err := scheduler.c.AddFunc(rule, func() {
			task.Execute() //根据rule定时执行这个任务
		})

		if err != nil {
			return err
		}
		scheduler.tasks[entryId] = task
	}
	return nil
}

func (scheduler *TaskScheduler) RemoveContainer(containerId string) error {
	scheduler.Lock()
	defer scheduler.Unlock()
	for entryId, task := range scheduler.tasks {
		if task.ContainerId == containerId {
			scheduler.c.Remove(entryId)
			delete(scheduler.tasks, entryId)
		}
	}
	return nil
}
