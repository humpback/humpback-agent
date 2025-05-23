package schedule

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

type Task struct {
	ContainerId string
	Name        string
	Image       string
	AlwaysPull  bool
	Timeout     time.Duration
	Rule        string
	client      *client.Client
	executing   bool
	Auth        string
}

func NewTask(containerId string, name string, image string, alwaysPull bool, timeout time.Duration, rule string, authStr string, client *client.Client) *Task {
	logrus.Infof("container %s task [%s] created.", name, rule)
	return &Task{
		ContainerId: containerId,
		Name:        name,
		Image:       image,
		AlwaysPull:  alwaysPull,
		Timeout:     timeout,
		Rule:        rule,
		client:      client,
		executing:   false,
		Auth:        authStr,
	}
}

func (task *Task) Execute() {
	if task.executing {
		logrus.Warnf("container %s task [%s] currently executing", task.Name, task.Rule)
		return
	}

	logrus.Infof("container %s task [%s] start executing", task.Name, task.Rule)

	task.executing = true
	reCreate := false
	if task.AlwaysPull { //检查镜像是否需要重启拉取
		currentImageId, err := task.getImageId()
		if err != nil {
			task.executing = false
			return
		}

		newImageId, err := task.pullImage()
		if err != nil {
			logrus.Errorf("container %s task [%s] pull image execute error, %v", task.Name, task.Rule, err)
			task.executing = false
			return
		}

		if currentImageId != newImageId {
			reCreate = true
		}
	}
	if task.Timeout <= 0 {
		if err := task.startContainer(context.Background(), reCreate); err != nil {
			logrus.Errorf("container %s task [%s] start container execute error, %v", task.Name, task.Rule, err)
		}
		task.executing = false
		return
	}

	// 设置任务的最大执行时间（超时时间）
	ctx, cancel := context.WithTimeout(context.Background(), task.Timeout)
	defer func() {
		cancel()
		task.executing = false
	}()

	// 启动容器
	if err := task.startContainer(ctx, reCreate); err != nil {
		logrus.Errorf("container %s task [%s] start container execute error, %v", task.Name, task.Rule, err)
		return
	}

	logrus.Infof("container %s task [%s] start succeed.", task.Name, task.Rule)
	select {
	case <-task.waitForContainerExit(ctx): // 等待容器完成任务
		logrus.Infof("container %s task [%s] executed.", task.Name, task.Rule)
	case <-ctx.Done(): // 容器执行超时
		logrus.Infof("container %s task [%s] executing timeout.", task.Name, task.Rule)
		task.stopContainer()
	}
}

func (task *Task) getImageId() (string, error) {
	imageInfo, _, err := task.client.ImageInspectWithRaw(context.Background(), task.Image)
	if err != nil {
		return "", err
	}
	return imageInfo.ID, nil
}

func (task *Task) pullImage() (string, error) {

	authStr := ""
	if task.Auth != "" {

		decodedBytes, err := base64.StdEncoding.DecodeString(task.Auth)
		if err != nil {
			return "", err
		}
		decodedStr := string(decodedBytes)

		// 按 ^^ 分割字符串
		parts := strings.Split(decodedStr, "^^")
		if len(parts) != 3 {
			return "", errors.New("image auth config invalid")
		}

		username := parts[0]
		password := parts[1]
		address := parts[2]

		authConfig := registry.AuthConfig{
			Username:      username,
			Password:      password,
			ServerAddress: address,
		}

		authBytes, _ := json.Marshal(authConfig)
		authStr = base64.URLEncoding.EncodeToString(authBytes)
	}

	pullOptions := image.PullOptions{
		All:          false,
		RegistryAuth: authStr,
	}

	out, err := task.client.ImagePull(context.Background(), task.Image, pullOptions)
	if err != nil {
		return "", err
	}

	out.Close()
	return task.getImageId()
}

func (task *Task) startContainer(ctx context.Context, reCreate bool) error {
	if reCreate {
		if err := task.reCreateContainer(); err != nil {
			return err
		}
	}
	return task.client.ContainerStart(ctx, task.ContainerId, container.StartOptions{})
}

func (task *Task) waitForContainerExit(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		statusCh, errCh := task.client.ContainerWait(ctx, task.ContainerId, container.WaitConditionNotRunning)
		select {
		case err := <-errCh:
			logrus.Errorf("container %s task [%s] exit error: %v", task.Name, task.Rule, err)
		case <-statusCh:
			// 容器已退出
		}
	}()
	return done
}

func (task *Task) stopContainer() error {
	timeout := MaxKillContainerTimeoutSeconds // 停止容器的超时时间, sdk单位为秒
	err := task.client.ContainerStop(context.Background(), task.ContainerId, container.StopOptions{
		Signal:  "SIGKILL",
		Timeout: &timeout,
	})

	if err != nil {
		logrus.Errorf("container %s task [%s] stop error: %s", task.Name, task.Rule, err.Error())
	}
	return err
}

func (task *Task) reCreateContainer() error {
	originContainerInfo, err := task.client.ContainerInspect(context.Background(), task.ContainerId)
	if err != nil {
		return err
	}

	discardContainerName := fmt.Sprintf("%s-%d-discard", originContainerInfo.Name, time.Now().Unix())
	//先将当前容器名称修改为废弃名称
	if err := task.client.ContainerRename(context.Background(), task.ContainerId, discardContainerName); err != nil {
		logrus.Errorf("container %s rename to %s error, always pull recreate give up. %s.", originContainerInfo.Name, discardContainerName, err.Error())
		return err
	}

	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: originContainerInfo.NetworkSettings.Networks,
	}

	containerInfo, err := task.client.ContainerCreate(context.Background(), originContainerInfo.Config, originContainerInfo.HostConfig, networkingConfig, nil, originContainerInfo.Name)
	if err != nil {
		logrus.Errorf("container %s always pull recreate error, %s.", originContainerInfo.Name, err.Error())
		task.client.ContainerRename(context.Background(), task.ContainerId, originContainerInfo.Name) //老容器还原名称
		return err
	}

	//删除老容器
	task.ContainerId = containerInfo.ID
	task.client.ContainerRemove(context.Background(), originContainerInfo.ID, container.RemoveOptions{Force: true})
	logrus.Infof("container %s recreated succeed.", originContainerInfo.Name)
	return nil
}
