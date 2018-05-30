package controllers

import "github.com/astaxie/beego"
import "github.com/docker/docker/api/types"
import "github.com/docker/docker/api/types/filters"
import "github.com/docker/libcompose/docker"
import "github.com/docker/libcompose/docker/ctx"
import "github.com/docker/libcompose/labels"
import "github.com/docker/libcompose/project"
import "github.com/docker/libcompose/project/options"
import "github.com/humpback/common/models"

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
)

var initialFilterArgs = []filters.KeyValuePair{
	filters.KeyValuePair{"label", labels.PROJECT.Str()},
	filters.KeyValuePair{"label", labels.SERVICE.Str()},
	filters.KeyValuePair{"label", labels.NUMBER.Str()},
}

// ServiceMaxTimeoutSecond - Service stop | restart max timeout
const ServiceMaxTimeoutSecond = 300

// ProjectPair - project service containers
type ProjectPair struct {
	Name       string
	Containers project.InfoSet
}

// ServiceController - handle http request for compose service
type ServiceController struct {
	baseController
	ComposeStorage *models.ComposeStorage
}

func createComposeAPIProject(projectName string, composeFile string) (project.APIProject, error) {

	composeAPIProject, err := docker.NewProject(&ctx.Context{
		Context: project.Context{
			ComposeFiles: []string{composeFile},
			ProjectName:  projectName,
		},
	}, nil)

	if err != nil {
		return nil, err
	}
	return composeAPIProject, nil
}

func containerName(names []string) string {

	max := math.MaxInt32
	var current string
	for _, v := range names {
		if len(v) < max {
			max = len(v)
			current = v
		}
	}
	return current[1:]
}

func containerPortString(ports []types.Port) string {

	result := []string{}
	for _, port := range ports {
		if port.PublicPort > 0 {
			result = append(result, fmt.Sprintf("%s:%d->%d/%s", port.IP, port.PublicPort, port.PrivatePort, port.Type))
		} else {
			result = append(result, fmt.Sprintf("%d/%s", port.PrivatePort, port.Type))
		}
	}
	return strings.Join(result, ", ")
}

func containerInfo(container types.Container) project.Info {

	result := project.Info{}
	result["Id"] = container.ID
	result["Name"] = containerName(container.Names)
	result["Command"] = container.Command
	result["State"] = container.Status
	result["Ports"] = containerPortString(container.Ports)
	return result
}

func (controller *ServiceController) GetServices() {

	filtersArgs := filters.NewArgs(initialFilterArgs...)
	option := types.ContainerListOptions{
		All:     true,
		Filters: filtersArgs,
	}

	containers, err := dockerClient.ContainerList(context.Background(), option)
	if err != nil {
		controller.Error(500, err.Error())
	}

	projectsKVPair := map[string]*ProjectPair{}
	for _, container := range containers {
		projectName := container.Labels[labels.PROJECT.Str()]
		info := containerInfo(container)
		if projectPair, _ := projectsKVPair[projectName]; projectPair == nil {
			projectsKVPair[projectName] = &ProjectPair{
				Name:       projectName,
				Containers: project.InfoSet{info},
			}
		} else {
			projectPair.Containers = append(projectPair.Containers, info)
		}
	}

	projectsBase := []*models.ProjectBase{}
	for _, projectPair := range projectsKVPair {
		projectBase := &models.ProjectBase{
			Name:       projectPair.Name,
			Containers: projectPair.Containers,
		}
		projectJSON, err := controller.ComposeStorage.ProjectJSON(projectPair.Name)
		if err != nil {
			projectStr := fmt.Sprintf("%s-%s-%d", projectPair.Name, "", 0)
			projectBase.HashCode = fmt.Sprintf("%x", sha256.Sum256([]byte(projectStr)))
			projectBase.Timestamp = 0
		} else {
			projectBase.HashCode = projectJSON.HashCode
			projectBase.Timestamp = projectJSON.Timestamp
		}
		projectsBase = append(projectsBase, projectBase)
	}
	controller.JSON(projectsBase)
}

func (controller *ServiceController) GetService() {

	projectName := controller.Ctx.Input.Param(":service")
	filtersArgs := filters.NewArgs(initialFilterArgs...)
	filtersArgs.Add("label", labels.PROJECT.Str()+"="+projectName)
	option := types.ContainerListOptions{
		All:     true,
		Filters: filtersArgs,
	}

	containers, err := dockerClient.ContainerList(context.Background(), option)
	if err != nil {
		controller.Error(500, err.Error())
	}

	if len(containers) == 0 {
		_, err := controller.ComposeStorage.ProjectJSON(projectName)
		if err == os.ErrNotExist {
			controller.Error(404, "service not found")
		}
	}

	projectInfo := &models.ProjectInfo{
		ProjectBase: models.ProjectBase{
			Name:       projectName,
			Containers: project.InfoSet{},
		},
	}

	for _, container := range containers {
		info := containerInfo(container)
		projectInfo.Containers = append(projectInfo.Containers, info)
	}

	projectData, err := controller.ComposeStorage.ProjectSpec(projectName)
	if err != nil {
		projectStr := fmt.Sprintf("%s-%s-%d", projectName, "", 0)
		projectInfo.HashCode = fmt.Sprintf("%x", sha256.Sum256([]byte(projectStr)))
		projectInfo.Timestamp = 0
	} else {
		projectInfo.HashCode = projectData.HashCode
		projectInfo.Timestamp = projectData.Timestamp
		projectInfo.ComposeData = string(projectData.ComposeBytes)
		projectInfo.PackageFile = projectData.PackageFile
	}
	controller.JSON(projectInfo)
}

func (controller *ServiceController) CreateService() {

	var reqBody models.CreateProject
	json.Unmarshal(controller.Ctx.Input.RequestBody, &reqBody)
	projectData, err := controller.ComposeStorage.CreateProjectSpec(reqBody)
	if err != nil {
		if err == os.ErrExist {
			controller.Error(409, "service already exists")
		}
		controller.Error(500, err.Error())
	}

	composeAPIProject, err := createComposeAPIProject(projectData.Name, projectData.ComposeFile)
	if err != nil {
		controller.ComposeStorage.RemoveProjectSpec(projectData.Name)
		if strings.Index(err.Error(), "yaml: unmarshal errors") >= 0 {
			err = fmt.Errorf("compose data format invalid, %s", err)
		}
		controller.Error(500, err.Error())
	}

	if err = composeAPIProject.Up(context.Background(), options.Up{}); err != nil {
		controller.Error(500, err.Error())
	}

	controller.JSON(&models.ProjectStatus{
		Name:   projectData.Name,
		Status: "services up.",
	})
}

func (controller *ServiceController) DeleteService() {

	projectName := controller.Ctx.Input.Param(":service")
	projectData, err := controller.ComposeStorage.ProjectSpec(projectName)
	if err != nil {
		if err == os.ErrNotExist {
			controller.Error(404, "service not found")
		}
		controller.Error(500, err.Error())
	}

	composeAPIProject, err := createComposeAPIProject(projectData.Name, projectData.ComposeFile)
	if err != nil {
		controller.Error(500, err.Error())
	}

	if err = composeAPIProject.Down(context.Background(), options.Down{}); err != nil {
		controller.Error(500, err.Error())
	}

	controller.ComposeStorage.RemoveProjectSpec(projectData.Name)
	controller.JSON(&models.ProjectStatus{
		Name:   projectData.Name,
		Status: "services down.",
	})
}

func (controller *ServiceController) OperateService() {

	var reqBody models.OperateProject
	json.Unmarshal(controller.Ctx.Input.RequestBody, &reqBody)
	projectData, err := controller.ComposeStorage.ProjectSpec(reqBody.Name)
	if err != nil {
		if err == os.ErrNotExist {
			controller.Error(404, "service not found")
		}
		controller.Error(500, err.Error())
	}

	composeAPIProject, err := createComposeAPIProject(projectData.Name, projectData.ComposeFile)
	if err != nil {
		controller.Error(500, err.Error())
	}

	infoSet, err := composeAPIProject.Ps(context.Background())
	if err != nil {
		controller.Error(500, err.Error())
	}

	servicesState := make(map[string]*models.ServiceState)
	for _, info := range infoSet {
		servicesState[info["Id"]] = &models.ServiceState{}
	}

	for containerid := range servicesState {
		container, err := dockerClient.ContainerInspect(context.Background(), containerid)
		if err == nil {
			servicesState[containerid].Name = container.Config.Labels["com.docker.compose.service"]
			servicesState[containerid].ContainerState = container.State
		}
	}

	beego.Info("compose project " + projectData.Name + " " + reqBody.Action)
	switch reqBody.Action {
	case "restart":
		beego.Info("restart " + projectData.Name)
		if err = composeAPIProject.Restart(context.Background(), ServiceMaxTimeoutSecond); err != nil {
			controller.Error(500, err.Error())
		}
	case "start":
		services := []string{}
		for _, serviceState := range servicesState {
			if serviceState.Name != "" && serviceState.ContainerState != nil {
				state := models.ContainerStateString(serviceState.ContainerState)
				if state == "Exited" || state == "Dead" || state == "Created" {
					services = append(services, serviceState.Name)
				}
			}
		}

		if len(services) > 0 {
			beego.Info("start " + strings.Join(services, ","))
			if err = composeAPIProject.Start(context.Background(), services...); err != nil {
				controller.Error(500, err.Error())
			}
		}
	case "stop":
		services := []string{}
		for _, serviceState := range servicesState {
			if serviceState.Name != "" && serviceState.ContainerState != nil {
				state := models.ContainerStateString(serviceState.ContainerState)
				if state == "Running" || state == "Restarting" || state == "Paused" {
					services = append(services, serviceState.Name)
				}
			}
		}
		if len(services) > 0 {
			beego.Info("stop " + strings.Join(services, ","))
			if err = composeAPIProject.Stop(context.Background(), ServiceMaxTimeoutSecond, services...); err != nil {
				controller.Error(500, err.Error())
			}
		}
	case "pause":
		services := []string{}
		for _, serviceState := range servicesState {
			if serviceState.Name != "" && serviceState.ContainerState != nil {
				state := models.ContainerStateString(serviceState.ContainerState)
				if state == "Running" {
					services = append(services, serviceState.Name)
				}
			}
		}
		if len(services) > 0 {
			beego.Info("pause " + strings.Join(services, ","))
			if err = composeAPIProject.Pause(context.Background(), services...); err != nil {
				controller.Error(500, err.Error())
			}
		}
	case "unpause":
		services := []string{}
		for _, serviceState := range servicesState {
			if serviceState.Name != "" && serviceState.ContainerState != nil {
				state := models.ContainerStateString(serviceState.ContainerState)
				if state == "Paused" {
					services = append(services, serviceState.Name)
				}
			}
		}
		if len(services) > 0 {
			beego.Info("unpause " + strings.Join(services, ","))
			if err = composeAPIProject.Unpause(context.Background(), services...); err != nil {
				controller.Error(500, err.Error())
			}
		}
	case "kill":
		services := []string{}
		for _, serviceState := range servicesState {
			if serviceState.Name != "" && serviceState.ContainerState != nil {
				state := models.ContainerStateString(serviceState.ContainerState)
				if state == "Running" || state == "Paused" {
					services = append(services, serviceState.Name)
				}
			}
		}
		if len(services) > 0 {
			beego.Info("kill " + strings.Join(services, ","))
			if err = composeAPIProject.Kill(context.Background(), "SIGKILL", services...); err != nil {
				controller.Error(500, err.Error())
			}
		}
	}

	controller.JSON(&models.ProjectStatus{
		Name:   projectData.Name,
		Status: "services " + reqBody.Action + ".",
	})
}

func (controller *ServiceController) PackageUploadService() {

	projectName := controller.Ctx.Input.Param(":service")
	mimeType := controller.Ctx.Input.Header("Content-Type")
	if mimeType == "" || strings.ToLower(mimeType) != "application/tar" {
		controller.Error(415, "package file format invalid")
	}

	packageFileBytes := controller.Ctx.Input.RequestBody
	if len(packageFileBytes) == 0 {
		controller.Error(400, "request body invalid")
	}

	var err error
	packageFile := controller.GetString("filename")
	packageFile, err = controller.ComposeStorage.SaveProjectPackageFile(projectName, packageFile, packageFileBytes)
	if err != nil {
		controller.Error(500, err.Error())
	}

	controller.JSON(&models.ProjectUploadStatus{
		Name:        projectName,
		PackageFile: packageFile,
		Status:      "service package upload successed.",
	})
}
