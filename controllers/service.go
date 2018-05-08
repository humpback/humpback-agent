package controllers

import "github.com/astaxie/beego"
import "github.com/docker/libcompose/docker"
import "github.com/docker/libcompose/docker/ctx"
import "github.com/docker/libcompose/project"
import "github.com/docker/libcompose/project/options"
import "github.com/humpback/common/models"

import (
	"context"
	"encoding/json"
	"os"
	"strings"
)

// ServiceMaxTimeoutSecond - Service stop | restart max timeout
const ServiceMaxTimeoutSecond = 300

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

func (controller *ServiceController) GetServices() {

	projectDataArray, err := controller.ComposeStorage.ProjectSpecs()
	if err != nil {
		controller.Error(500, err.Error())
	}

	projectConfigs := []*models.ProjectConfig{}
	for _, projectData := range projectDataArray {
		composeAPIProject, err := createComposeAPIProject(projectData.Name, projectData.ComposeFile)
		if err != nil {
			controller.Error(500, err.Error())
		}
		infoSet, err := composeAPIProject.Ps(context.Background())
		if err != nil {
			controller.Error(500, err.Error())
		}
		projectConfigs = append(projectConfigs, &models.ProjectConfig{
			Name:        projectData.Name,
			HashCode:    projectData.HashCode,
			ComposeData: string(projectData.ComposeBytes),
			PackageFile: projectData.PackageFile,
			Timestamp:   projectData.Timestamp,
			Containers:  infoSet,
		})
	}
	controller.JSON(projectConfigs)
}

func (controller *ServiceController) GetService() {

	projectName := controller.Ctx.Input.Param(":service")
	projectData, err := controller.ComposeStorage.ProjectSpec(projectName)
	if err != nil {
		if err == os.ErrNotExist {
			controller.Error(404, err.Error())
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

	controller.JSON(&models.ProjectConfig{
		Name:        projectData.Name,
		HashCode:    projectData.HashCode,
		ComposeData: string(projectData.ComposeBytes),
		PackageFile: projectData.PackageFile,
		Timestamp:   projectData.Timestamp,
		Containers:  infoSet,
	})
}

func (controller *ServiceController) CreateService() {

	var reqBody models.CreateProject
	json.Unmarshal(controller.Ctx.Input.RequestBody, &reqBody)
	projectData, err := controller.ComposeStorage.CreateProjectSpec(reqBody)
	if err != nil {
		if err == os.ErrExist {
			controller.Error(409, err.Error())
		}
		controller.Error(500, err.Error())
	}

	composeAPIProject, err := createComposeAPIProject(projectData.Name, projectData.ComposeFile)
	if err != nil {
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
			controller.Error(404, err.Error())
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
			controller.Error(404, err.Error())
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
		beego.Info("restart...")
		if err = composeAPIProject.Restart(context.Background(), ServiceMaxTimeoutSecond); err != nil {
			controller.Error(500, err.Error())
		}
	case "start":
		//start, 只处理状态为：Exited | Dead | Created 的容器
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
		//stop, 只处理状态为：Running | Restarting | Paused 的容器
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
		//pause, 只处理状态为：Running的容器
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
		//unpause, 只处理状态为：Paused的容器
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
		//kill, 只处理状态为：Running | Paused 的容器
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
