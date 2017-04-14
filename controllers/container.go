package controllers

import (
	"common/models"
	"encoding/json"
	"fmt"
	"humpback-agent/config"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	gonetwork "github.com/humpback/gounits/network"
	"golang.org/x/net/context"
)

// ContainerController - handle http request for container
type ContainerController struct {
	baseController
}

var containerID string

// Prepare - format path before exec real action
func (ctCtrl *ContainerController) Prepare() {
	containerID = ctCtrl.Ctx.Input.Param(":containerid")
}

// GetContainer - get container info with name or id
func (ctCtrl *ContainerController) GetContainer() {
	var container models.Container
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
	res, err := dockerClient.ContainerLogs(context.Background(), containerID, option)
	if err != nil {
		ctCtrl.Error(500, err.Error())
	}
	defer res.Close()

	stdout, _ := ioutil.ReadAll(res)
	result := strings.Split(string(stdout), "\n")
	result = result[0 : len(result)-1]
	for i, item := range result {
		result[i] = string([]byte(item)[8:])
	}
	ctCtrl.JSON(result)
}

// GetContainerStats - get container's stats, include (cpu/memory/network...)
func (ctCtrl *ContainerController) GetContainerStats() {
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

	res, err := dockerClient.ContainerStats(context.Background(), containerID, false)
	if err != nil {
		if strings.Index(err.Error(), "No such container") != -1 {
			ctCtrl.Error(404, err.Error())
		} else {
			ctCtrl.Error(500, err.Error())
		}
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	resData, readIOErr := ioutil.ReadAll(res.Body)
	if readIOErr != nil {
		ctCtrl.Error(500, err.Error())
	}
	stats := models.ContainerStatsFromDocker{}

	if err := json.Unmarshal(resData, &stats); err != nil {
		ctCtrl.Error(500, err.Error())
	}

	containerStats := models.ContainerStats{}
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

	ctCtrl.JSON(containerStats)
}

// CreateContainer - create container
func (ctCtrl *ContainerController) CreateContainer() {
	var reqBody models.Container
	json.Unmarshal(ctCtrl.Ctx.Input.RequestBody, &reqBody)

	// determine whether there is a image
	err := tryPullImage(reqBody.Image)
	if err != nil {
		ctCtrl.Error(500, err.Error(), 21001)
	}

	config := container.Config{
		Hostname:     reqBody.HostName,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Env:          reqBody.Env,
		Image:        reqBody.Image,
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
	hostconfig.PortBindings = portBinding

	//volume binding
	config.Volumes = make(map[string]struct{})
	for _, item := range reqBody.Volumes {
		config.Volumes[item.ContainerVolume] = *new(struct{})
		tempBind := item.HostVolume + ":" + item.ContainerVolume
		hostconfig.Binds = append(hostconfig.Binds, tempBind)
	}

	res, err := dockerClient.ContainerCreate(context.Background(), &config, &hostconfig, &network.NetworkingConfig{}, reqBody.Name)
	if err != nil {
		if err != nil {
			if err.Error() == "container already exists" {
				ctCtrl.Error(409, err.Error())
			} else {
				ctCtrl.Error(500, err.Error())
			}
		}
	}

	err = dockerClient.ContainerStart(context.Background(), res.ID, types.ContainerStartOptions{})
	if err != nil {
		dockerClient.ContainerRemove(context.Background(), res.ID, types.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		})
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
		err = dockerClient.ContainerStop(context.Background(), reqBody.Container, &duration)
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
	removeOptions := types.ContainerRemoveOptions{
		RemoveVolumes: true,
	}
	force, _ := ctCtrl.GetBool("force", false)
	removeOptions.Force = force
	err := dockerClient.ContainerRemove(context.Background(), containerID, removeOptions)
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
	err = tryPullImage(newImage)
	fmt.Println(err)
	if err != nil {
		return "", 20004, err
	}

	beego.Debug("UPGRADE - Begin to reanme container from " + originalName + " to " + tempName)
	err = dockerClient.ContainerRename(context.Background(), id, tempName)
	if err != nil {
		return "", 20003, err
	}

	if container.State.Running {
		duration := time.Second * 10
		beego.Debug("UPGRADE - Begin to stop container info for " + id)
		if err := dockerClient.ContainerStop(context.Background(), id, &duration); err != nil {
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

func (ctCtrl *ContainerController) makeSystemIdlePort(kind string) int {

	var (
		err     error
		minPort int
		maxPort int
	)

	conf := config.GetConfig()
	if !conf.DockerClusterEnabled {
		return 0
	}

	rangePorts := strings.SplitN(conf.DockerClusterPortsRange, "-", 2)
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
