package schedule

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"time"
)

type Task struct {
	ContainerId string
	Image       string
	AlwaysPull  bool
	Timeout     time.Duration
	Rule        string
	client      *client.Client
}

func NewTask(containerId string, image string, alwaysPull bool, timeout time.Duration, rule string, client *client.Client) *Task {
	return &Task{
		ContainerId: containerId,
		Image:       image,
		AlwaysPull:  alwaysPull,
		Timeout:     timeout,
		Rule:        rule,
		client:      client,
	}
}

func (task *Task) Execute() {
	// 设置任务的最大执行时间（超时时间）
	ctx, cancel := context.WithTimeout(context.Background(), task.Timeout)
	defer cancel()

	// 启动容器
	if err := task.startContainer(ctx, task.client, task.ContainerId); err != nil {
		logrus.Errorf("container %s task [%s] start container execute error, %v", task.ContainerId, task.Rule, err)
		return
	}

	logrus.Infof("container %s task [%s] start succeed.", task.ContainerId, task.Rule)
	select {
	case <-task.waitForContainerExit(ctx, task.client, task.ContainerId): // 等待容器完成任务
		logrus.Infof("container %s task [%s] executed.", task.ContainerId, task.Rule)
	case <-ctx.Done(): // 容器执行超时
		logrus.Infof("container %s task [%s] executing timeout.", task.ContainerId, task.Rule)
		task.stopContainer(ctx, task.client, task.ContainerId)
	}
}

func (task *Task) pullImage() error {
	pullOptions := image.PullOptions{
		All: false,
	}

	out, err := task.client.ImagePull(context.Background(), task.Image, pullOptions)
	if err != nil {
		return err
	}

	out.Close()
	return nil
}

func (task *Task) startContainer(ctx context.Context, cli *client.Client, imageName string) error {
	if task.AlwaysPull {
		if err := task.pullImage(); err != nil {
			return err
		}
	}
	return cli.ContainerStart(ctx, task.ContainerId, container.StartOptions{})
}

func (task *Task) waitForContainerExit(ctx context.Context, cli *client.Client, containerID string) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		statusCh, errCh := cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
		select {
		case err := <-errCh:
			logrus.Errorf("container %s task [%s] exit error: %v", containerID, task.Rule, err)
		case <-statusCh:
			// 容器已退出
		}
	}()
	return done
}

func (task *Task) stopContainer(ctx context.Context, cli *client.Client, containerID string) error {
	timeout := 5 // 停止容器的超时时间, sdk单位为秒
	return cli.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout})
}
