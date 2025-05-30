package handler

import (
	"context"
	"fmt"
	"humpback-agent/api/factory"
	v1model "humpback-agent/api/v1/model"
	"humpback-agent/controller"
	"humpback-agent/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	APIVersion = "/v1"
)

func init() {
	_ = factory.InjectHandler(APIVersion, NewV1Handler)
}

type V1TaskType int

const (
	ContainerCreateTask  V1TaskType = 1
	ContainerDeleteTask  V1TaskType = 2
	ContainerStartTask   V1TaskType = 3
	ContainerStopTask    V1TaskType = 4
	ContainerRestartTask V1TaskType = 5
	NetworkCreateTask    V1TaskType = 6
	NetworkDeleteTask    V1TaskType = 7
)

type V1Task struct {
	TaskType V1TaskType
	TaskBody any
}

type V1Handler struct {
	factory.BaseHandler
	apiVersion string
	taskChan   chan *V1Task
}

func NewV1Handler(controller controller.ControllerInterface) factory.HandlerInterface {
	return &V1Handler{
		BaseHandler: factory.BaseHandler{
			ControllerInterface: controller,
		},
		taskChan: make(chan *V1Task),
	}
}

func (handler *V1Handler) watchTasks() {

	for task := range handler.taskChan {
		switch task.TaskType {
		case ContainerCreateTask:
			container := handler.Container()
			go container.Create(context.Background(), task.TaskBody.(*v1model.CreateContainerRequest))
		case ContainerDeleteTask:
			container := handler.Container()
			request := task.TaskBody.(*v1model.DeleteContainerRequest)
			containerMeta := model.ContainerMeta{
				ContainerName: request.ContainerName,
				IsDelete:      true,
			}
			container.BaseController().FailureChan() <- containerMeta
			go container.Delete(context.Background(), request)
		case ContainerRestartTask:
			container := handler.Container()
			go container.Restart(context.Background(), task.TaskBody.(*v1model.RestartContainerRequest))
		case ContainerStartTask:
			container := handler.Container()
			go container.Start(context.Background(), task.TaskBody.(*v1model.StartContainerRequest))
		case ContainerStopTask:
			container := handler.Container()
			go container.Stop(context.Background(), task.TaskBody.(*v1model.StopContainerRequest))
		case NetworkCreateTask:
			network := handler.Network()
			go network.Create(context.Background(), task.TaskBody.(*v1model.CreateNetworkRequest))
		case NetworkDeleteTask:
			network := handler.Network()
			go network.Delete(context.Background(), task.TaskBody.(*v1model.DeleteNetworkRequest))
		}
	}
}

func (handler *V1Handler) SetRouter(version string, engine *gin.Engine) {
	handler.apiVersion = version
	routerRouter := engine.Group(fmt.Sprintf("api/%s", handler.apiVersion))
	{
		routerRouter.GET("/faq", handler.faqHandleFunc)
		//container router
		containerRouter := routerRouter.Group("container")
		{
			containerRouter.GET(":containerId", handler.GetContainerHandleFunc)
			containerRouter.POST("list", handler.QueryContainerHandleFunc)
			containerRouter.POST("", handler.CreateContainerHandleFunc)
			containerRouter.PUT("", handler.UpdateContainerHandleFunc)
			containerRouter.DELETE(":containerId", handler.DeleteContainerHandleFunc)
			containerRouter.POST(":containerId/restart", handler.RestartContainerHandleFunc)
			containerRouter.POST(":containerId/start", handler.StartContainerHandleFunc)
			containerRouter.POST(":containerId/stop", handler.StopContainerHandleFunc)
			containerRouter.GET(":containerId/logs", handler.GetContainerLogsHandleFunc)
			containerRouter.GET(":containerId/stats", handler.GetContainerStatsHandleFunc)
		}

		//image router
		imageRouter := routerRouter.Group("image")
		{
			imageRouter.GET(":imageId", handler.GetImageHandleFunc)
			imageRouter.POST("query", handler.QueryImageHandleFunc)
			imageRouter.POST("push", handler.PushImageHandleFunc)
			imageRouter.POST("pull", handler.PullImageHandleFunc)
			imageRouter.DELETE(":imageId", handler.DeleteImageHandleFunc)
		}

		//volume router
		volumeRouter := routerRouter.Group("volume")
		{
			volumeRouter.GET(":volumeId", handler.GetVolumeHandleFunc)
			volumeRouter.POST("query", handler.QueryVolumeHandleFunc)
			volumeRouter.POST("", handler.CreateVolumeHandleFunc)
			volumeRouter.PUT("", handler.UpdateVolumeHandleFunc)
			volumeRouter.DELETE(":volumeId", handler.DeleteVolumeHandleFunc)
		}

		//network router
		networkRouter := routerRouter.Group("network")
		{
			networkRouter.GET(":networkId", handler.GetNetworkHandleFunc)
			networkRouter.POST("query", handler.QueryNetworkHandleFunc)
			networkRouter.POST("", handler.CreateNetworkHandleFunc)
			networkRouter.PUT("", handler.UpdateNetworkHandleFunc)
			networkRouter.DELETE(":networkId", handler.DeleteNetworkHandleFunc)
		}
	}
	go handler.watchTasks()
}

func (handler *V1Handler) faqHandleFunc(c *gin.Context) {
	c.JSON(http.StatusOK, &v1model.FAQResponse{
		APIVersion: handler.apiVersion,
		Timestamp:  time.Now().UnixMilli(),
	})
}
