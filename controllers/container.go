package controllers

import "github.com/humpback/common/models"
import "humpback-agent/config"
import gonetwork "github.com/humpback/gounits/network"
import "github.com/astaxie/beego"
import "github.com/docker/docker/api/types"
import "github.com/docker/docker/api/types/container"
import "github.com/docker/docker/api/types/network"
import "github.com/docker/go-connections/nat"

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
	"context"
)

// ContainerController - handle http request for container
type ContainerController struct {
	baseController
}

type containerError struct {
	Error error
}

var clientTimeout = 30 * time.Second

// Prepare - format path before exec real action
func (ctCtrl *ContainerController) Prepare() {
}

// GetContainer - get container info with name or id
func (ctCtrl *ContainerController) GetContainer() {
	var container models.Container
	containerID := ctCtrl.Ctx.Input.Param(":containerid")
	originalContainer, err := dockerClient.ContainerInspect(context.Background(), containerID)
	if err != nil {
		if strings.Index(err.Error(), "No such container") != -1 {
			ctCtrl.Error(404, err.Error())
		} else {
			ctCtrl.Error(500, err.Error())
		}
	}
	originaldata, _ := ctCtrl.GetBool("originaldata", false)
	if originaldata {
		ctCtrl.JSON(originalContainer)
	} else {
		container.Parse(&originalContainer)
		ctCtrl.JSON(container)
	}
}

// GetContainers - get containers
func (ctCtrl *ContainerController) GetContainers() {
	queryAll, _ := ctCtrl.GetBool("all", false)
	option := types.ContainerListOptions{
		All: queryAll,
	}
	containers, err := dockerClient.ContainerList(context.Background(), option)
	if err != nil {
		ctCtrl.Error(500, err.Error())
	}
	ctCtrl.JSON(containers)
}

// GetContainerLogs - get container
func (ctCtrl *ContainerController) GetContainerLogs() {
	tail := ctCtrl.GetString("tail")
	since := ctCtrl.Ctx.Input.Query("since")
	if since != "" {
		tempTime, err := time.Parse("2006-01-02", since)
		if err != nil {
			ctCtrl.Error(400, "Invalid date.")
		}
		since = strconv.FormatInt(tempTime.Unix(), 10)
	}

	option := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Since:      since,
		Timestamps: false,
	}
	if tail != "" {
		option.Tail = tail
	}

	containerID := ctCtrl.Ctx.Input.Param(":containerid")
	res, err := dockerClient.ContainerLogs(context.Background(), containerID, option)
	if err != nil {
		ctCtrl.Error(500, err.Error())
	}
	defer res.Close()

	stdout, _ := ioutil.ReadAll(res)
	result := strings.Split(string(stdout), "\n")
	result = result[0 : len(result)-1]
	ctCtrl.JSON(result)
}

// GetAllContainerStats - get all container's stats, include (cpu/memory/network...)
func (ctCtrl *ContainerController) GetAllContainerStats() {
	option := types.ContainerListOptions{
		All: false,
	}
	containers, err := dockerClient.ContainerList(context.Background(), option)
	if err != nil {
		ctCtrl.Error(500, err.Error())
	}

	chs := make([]chan models.ContainerStatsWithError, len(containers))
	stats := make([]models.ContainerStats, len(containers))
	for i, container := range containers {
		chs[i] = make(chan models.ContainerStatsWithError)
		go getContainerStats(container.Names[0], container.ID, chs[i])
	}

	for i, ch := range chs {
		value := <-ch
		stats[i] = value.Stats
	}

	ctCtrl.JSON(stats)
}

func getContainerStats(name string, id string, ch chan models.ContainerStatsWithError) {
	result := models.ContainerStatsWithError{}
	containerStats := models.ContainerStats{}
	containerStats.ContainerName = strings.Replace(name, "/", "", 1)
	containerStats.ContainerID = id

	result.Stats = containerStats
	res, err := dockerClient.ContainerStats(context.Background(), id, false)
	if err != nil {
		result.Error = err
		ch <- result
		return
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	resData, readIOErr := ioutil.ReadAll(res.Body)
	if readIOErr != nil {
		result.Error = err
		ch <- result
		return
	}
	stats := models.ContainerStatsFromDocker{}
	if err := json.Unmarshal(resData, &stats); err != nil {
		result.Error = err
		ch <- result
		return
	}

	containerStats.NetworkIn = int64(stats.Network.RxBytes / 1024)
	containerStats.NetworkOut = int64(stats.Network.TxBytes / 1024)
	containerStats.MemoryUsage = int64(stats.MemoryStats.Usage / 1024)
	containerStats.MemoryLimit = int64(stats.MemoryStats.Limit / 1024)
	containerStats.MemoryPercent = (stats.MemoryStats.Usage / stats.MemoryStats.Limit) * 100

	cpuDelta := stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage
	systemDelta := stats.CPUStats.SystemCPUUsage - stats.PreCPUStats.SystemCPUUsage
	resultCPUUsage := cpuDelta / systemDelta * 100
	containerStats.CPUUsage = fmt.Sprintf("%.2f", resultCPUUsage)

	for _, blk := range stats.BlkioStats.IOServiceBytesRecursive {
		if blk.Op == "Read" {
			containerStats.IOBytesRead = blk.Value
		}
		if blk.Op == "Write" {
			containerStats.IOBytesWrite = blk.Value
		}
	}
	result.Stats = containerStats
	ch <- result
}

// GetContainerStats - get container's stats, include (cpu/memory/network...)
func (ctCtrl *ContainerController) GetContainerStats() {
	containerID := ctCtrl.Ctx.Input.Param(":containerid")
	container, err := dockerClient.ContainerInspect(context.Background(), containerID)
	if err != nil {
		if strings.Index(err.Error(), "No such container") != -1 {
			ctCtrl.Error(404, err.Error())
		} else {
			ctCtrl.Error(500, err.Error())
		}
	}
	if !container.State.Running {
		ctCtrl.Error(500, "Container is not running")
	}

	ch := make(chan models.ContainerStatsWithError)
	go getContainerStats(container.Name, container.ID, ch)
	containerStatsWithError := <-ch
	if containerStatsWithError.Error != nil {
		ctCtrl.Error(500, containerStatsWithError.Error)
	} else {
		ctCtrl.JSON(containerStatsWithError.Stats)
	}
}

// GetContainerStatus - get container's status (running, paused, restarting, killed, dead, pid, exitcode...)
func (ctCtrl *ContainerController) GetContainerStatus() {
	containerID := ctCtrl.Ctx.Input.Param(":containerid")
	container, err := dockerClient.ContainerInspect(context.Background(), containerID)
	if err != nil {
		if strings.Index(err.Error(), "No such container") != -1 {
			ctCtrl.Error(404, err.Error())
		} else {
			ctCtrl.Error(500, err.Error())
		}
	}
	ctCtrl.JSON(container.State)
}

// CreateContainer - create container
func (ctCtrl *ContainerController) CreateContainer() {
	var reqBody models.Container
	json.Unmarshal(ctCtrl.Ctx.Input.RequestBody, &reqBody)

	procUpdate := struct {
		AllowUpdate       bool
		OriginalName      string
		OriginalContainer types.ContainerJSON
	}{
		AllowUpdate: false,
	}

	if strings.TrimSpace(reqBody.ID) != "" { //UpdateContainer Req
		originalContainer, err := dockerClient.ContainerInspect(context.Background(), reqBody.ID)
		if err != nil {
			if strings.Index(err.Error(), "No such container") != -1 {
				ctCtrl.Error(404, err.Error())
			} else {
				ctCtrl.Error(500, err.Error())
			}
		} else {
			procUpdate.AllowUpdate = true
			procUpdate.OriginalName = strings.Replace(originalContainer.Name, "/", "", 1)
			procUpdate.OriginalContainer = originalContainer
		}
	}

	// determine whether there is a image
	inspectImage, err := tryPullImage(reqBody.Image)
	if err != nil {
		ctCtrl.Error(500, err.Error(), 21001)
	}

	config := container.Config{
		Hostname:     reqBody.HostName,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		OpenStdin:    true,
		Tty:          true,
		Env:          reqBody.Env,
		Image:        reqBody.Image,
		Labels:       reqBody.Labels,
	}

	if reqBody.Command != "" {
		config.Cmd = strings.SplitN(reqBody.Command, " ", -1)
	}

	if strings.ToLower(reqBody.NetworkMode) == "host" {
		config.Hostname = ""
	}

	hostconfig := container.HostConfig{
		PublishAllPorts: true,
		NetworkMode:     container.NetworkMode(reqBody.NetworkMode),
		DNS:             reqBody.DNS,
		ExtraHosts:      reqBody.ExtraHosts,
		ShmSize:         reqBody.SHMSize,
		Links:           reqBody.Links,
		LogConfig:       reqBody.LogConfig,
	}

	if reqBody.Ulimits != nil {
		hostconfig.Ulimits = reqBody.Ulimits
	}

	hostconfig.CPUShares = reqBody.CPUShares
	hostconfig.Memory = reqBody.Memory * 1024 * 1024
	if hostconfig.ShmSize == 0 {
		hostconfig.ShmSize = 67108864
	}
	if reqBody.RestartPolicy != "" {
		hostconfig.RestartPolicy.Name = reqBody.RestartPolicy
		hostconfig.RestartPolicy.MaximumRetryCount = reqBody.RestartRetryCount
	}

	//port binding
	config.ExposedPorts = make(nat.PortSet)
	portBinding := make(nat.PortMap)
	if hostconfig.NetworkMode.IsBridge() {
		if len(reqBody.Ports) > 0 {
			hostconfig.PublishAllPorts = false //disable -P ports alloc.
			for _, item := range reqBody.Ports {
				privatePort := nat.Port(strconv.Itoa(item.PrivatePort) + "/" + item.Type)
				config.ExposedPorts[privatePort] = *new(struct{})
				if item.PublicPort != 0 {
					tempPublicPort := []nat.PortBinding{
						nat.PortBinding{
							HostIP:   item.IP,
							HostPort: strconv.Itoa(item.PublicPort),
						},
					}
					portBinding[privatePort] = tempPublicPort
				} else {
					hostPort := ctCtrl.makeSystemIdlePort(item.Type)
					if hostPort > 0 {
						tempPublicPort := []nat.PortBinding{
							nat.PortBinding{
								HostIP:   item.IP,
								HostPort: strconv.Itoa((int)(hostPort)),
							},
						}
						portBinding[privatePort] = tempPublicPort
					}
				}
			}
		} else {
			for item := range inspectImage.Config.ExposedPorts {
				privatePort := item
				config.ExposedPorts[privatePort] = *new(struct{})
				hostPort := ctCtrl.makeSystemIdlePort(item.Proto())
				if hostPort > 0 {
					tempPublicPort := []nat.PortBinding{
						nat.PortBinding{
							HostIP:   "0.0.0.0",
							HostPort: strconv.Itoa((int)(hostPort)),
						},
					}
					portBinding[privatePort] = tempPublicPort
				}
			}
		}
	}
	hostconfig.PortBindings = portBinding

	//volume binding
	config.Volumes = make(map[string]struct{})
	for _, item := range reqBody.Volumes {
		config.Volumes[item.ContainerVolume] = *new(struct{})
		tempBind := item.HostVolume + ":" + item.ContainerVolume
		hostconfig.Binds = append(hostconfig.Binds, tempBind)
	}

	if procUpdate.AllowUpdate {
		tempName := procUpdate.OriginalName + strconv.FormatInt(time.Now().Unix(), 10)
		beego.Debug("UPDATE - Begin to reanme container from " + procUpdate.OriginalName + " to " + tempName)
		err := dockerClient.ContainerRename(context.Background(), procUpdate.OriginalContainer.ID, tempName)
		if err != nil {
			ctCtrl.Error(500, err.Error(), 20003)
		}
		if procUpdate.OriginalContainer.State.Running || procUpdate.OriginalContainer.State.Restarting {
			beego.Debug("UPDATE - Begin to stop container info for " + procUpdate.OriginalContainer.ID)
			if err := stopContainer(procUpdate.OriginalContainer.ID); err != nil {
				dockerClient.ContainerRename(context.Background(), procUpdate.OriginalContainer.ID, procUpdate.OriginalName)
				ctCtrl.Error(500, err.Error(), 20005)
			}
		}
	}

	networkConfig := network.NetworkingConfig{}
	res, err := dockerClient.ContainerCreate(context.Background(), &config, &hostconfig, &networkConfig, reqBody.Name)
	if err != nil {
		if procUpdate.AllowUpdate {
			dockerClient.ContainerRename(context.Background(), procUpdate.OriginalContainer.ID, procUpdate.OriginalName)
		}
		if err.Error() == "container already exists" {
			ctCtrl.Error(409, err.Error())
		} else {
			ctCtrl.Error(500, err.Error())
		}
	}

	if procUpdate.AllowUpdate {
		beego.Debug("UPDATE - Begin to delete old container " + procUpdate.OriginalContainer.ID)
		dockerClient.ContainerRemove(context.Background(), procUpdate.OriginalContainer.ID, types.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		})
	}

	err = dockerClient.ContainerStart(context.Background(), res.ID, types.ContainerStartOptions{})
	if err != nil {
		if !procUpdate.AllowUpdate {
			dockerClient.ContainerRemove(context.Background(), res.ID, types.ContainerRemoveOptions{
				RemoveVolumes: true,
				Force:         true,
			})
		}
		ctCtrl.Error(500, err.Error(), 20001)
	}
	result := map[string]interface{}{
		"Id":       res.ID,
		"Name":     reqBody.Name,
		"Warnings": res.Warnings,
	}
	ctCtrl.JSON(result)
}

// OperateContainer - start/stop/restart...
func (ctCtrl *ContainerController) OperateContainer() {
	reqBody := models.ContainerOperate{}
	json.Unmarshal(ctCtrl.Ctx.Input.RequestBody, &reqBody)

	var err error
	var errCode int
	var newID string
	duration := time.Second * 10
	switch strings.ToLower(reqBody.Action) {
	case "start":
		err = dockerClient.ContainerStart(context.Background(), reqBody.Container, types.ContainerStartOptions{})
		break
	case "stop":
		err = stopContainer(reqBody.Container)
		break
	case "restart":
		err = dockerClient.ContainerRestart(context.Background(), reqBody.Container, &duration)
		break
	case "kill":
		err = dockerClient.ContainerKill(context.Background(), reqBody.Container, "SIGKILL")
		break
	case "pause":
		err = dockerClient.ContainerPause(context.Background(), reqBody.Container)
		break
	case "unpause":
		err = dockerClient.ContainerUnpause(context.Background(), reqBody.Container)
		break
	case "rename":
		err = dockerClient.ContainerRename(context.Background(), reqBody.Container, reqBody.NewName)
		break
	case "upgrade":
		newID, errCode, err = upgradeContainer(reqBody.Container, reqBody.ImageTag)
	}
	if err != nil {
		if strings.Index(err.Error(), "No such container") != -1 {
			ctCtrl.Error(404, err.Error())
		} else {
			ctCtrl.Error(500, err.Error(), errCode)
		}
	}
	if newID != "" {
		result := map[string]string{"Id": newID}
		ctCtrl.JSON(result)
	}
}

// DeleteContainer - delete container
func (ctCtrl *ContainerController) DeleteContainer() {

	containerID := ctCtrl.Ctx.Input.Param(":containerid")
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	removeCh := make(chan containerError, 1)
	var err error
	go func(container string, ch chan<- containerError) {
		removeError := containerError{Error: nil}
		removeOptions := types.ContainerRemoveOptions{
			RemoveVolumes: true,
		}
		force, _ := ctCtrl.GetBool("force", false)
		removeOptions.Force = force
		removeError.Error = dockerClient.ContainerRemove(context.Background(), containerID, removeOptions)
		ch <- removeError
	}(containerID, removeCh)
	select {
	case removeError := <-removeCh:
		err = removeError.Error
	case <-ctx.Done():
		err = fmt.Errorf("remove container %s timeout", containerID)
	}
	cancel()
	close(removeCh)
	if err != nil {
		if strings.Index(err.Error(), "No such container") != -1 {
			ctCtrl.Error(404, err.Error())
		} else {
			ctCtrl.Error(500, err.Error())
		}
	}
}

// upgradeContainer - upgrade containers's image
func upgradeContainer(id, newTag string) (string, int, error) {
	beego.Info("UPGRADE - Begin to get container info for " + id)
	container, err := dockerClient.ContainerInspect(context.Background(), id)
	if err != nil {
		return "", 20002, err
	}
	id = container.ID[0:12]

	originalName := strings.Replace(container.Name, "/", "", 1)
	tempName := originalName + strconv.FormatInt(time.Now().Unix(), 10)
	newImage := strings.Split(container.Config.Image, ":")[0] + ":" + newTag

	beego.Debug("UPGRADE - Begin try to pull image " + newImage)
	_, err = tryPullImage(newImage)
	if err != nil {
		return "", 20004, err
	}

	beego.Debug("UPGRADE - Begin to reanme container from " + originalName + " to " + tempName)
	err = dockerClient.ContainerRename(context.Background(), id, tempName)
	if err != nil {
		return "", 20003, err
	}

	if container.State.Running || container.State.Restarting {
		beego.Debug("UPGRADE - Begin to stop container info for " + id)
		if err := stopContainer(id); err != nil {
			return "", 20005, err
		}
	}

	config := *container.Config
	config.Image = newImage
	hostconfig := *container.HostConfig
	if hostconfig.NetworkMode.IsHost() {
		config.Hostname = ""
		config.Domainname = ""
	}

	beego.Debug("UPGRADE - Begin to create new container use image " + newImage + " for " + id)
	res, err := dockerClient.ContainerCreate(context.Background(), &config, &hostconfig, &network.NetworkingConfig{}, originalName)
	if err != nil {
		dockerClient.ContainerRename(context.Background(), id, originalName)
		return "", 20006, err
	}

	beego.Debug("UPGRADE - Begin to start new container " + res.ID)
	err = dockerClient.ContainerStart(context.Background(), res.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", 20007, err
	}

	removeOption := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}

	beego.Debug("UPGRADE - Begin to delete old container " + id)
	err = dockerClient.ContainerRemove(context.Background(), id, removeOption)
	if err != nil {
		return res.ID, 20008, err
	}
	return res.ID, 0, nil
}

func stopContainer(id string) error {

	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	stopCh := make(chan containerError, 1)
	var err error
	go func(container string, ch chan<- containerError) {
		stopError := containerError{Error: nil}
		duration := time.Second * 20
		stopError.Error = dockerClient.ContainerStop(context.Background(), container, &duration)
		ch <- stopError
	}(id, stopCh)
	select {
	case stopError := <-stopCh:
		err = stopError.Error
	case <-ctx.Done():
		err = fmt.Errorf("stop container %s timeout", id)
	}
	cancel()
	close(stopCh)
	return err
}

func (ctCtrl *ContainerController) makeSystemIdlePort(kind string) int {

	var (
		err     error
		minPort int
		maxPort int
	)

	conf := config.GetConfig()
	rangePorts := strings.SplitN(conf.DockerContainerPortsRange, "-", 2)
	if len(rangePorts) != 2 {
		return 0
	}

	if minPort, err = strconv.Atoi(rangePorts[0]); err != nil {
		return 0
	}

	if maxPort, err = strconv.Atoi(rangePorts[1]); err != nil {
		return 0
	}

	port, err := gonetwork.GrabSystemRangeIdlePort(kind, (uint32)(minPort), (uint32)(maxPort))
	if err != nil {
		return 0
	}
	return (int)(port)
}
