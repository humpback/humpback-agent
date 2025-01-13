package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"humpback-agent/pkg/api/factory"
	"humpback-agent/pkg/api/v1/model"
	"humpback-agent/pkg/controller"
	"net/http"
	"time"
)

const (
	APIVersion = "/v1"
)

func init() {
	_ = factory.InjectHandler(APIVersion, NewV1Handler)
}

type V1Handler struct {
	factory.BaseHandler
	apiVersion string
}

func NewV1Handler(controller controller.ControllerInterface) factory.HandlerInterface {
	return &V1Handler{
		BaseHandler: factory.BaseHandler{
			ControllerInterface: controller,
		},
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
}

func (handler *V1Handler) faqHandleFunc(c *gin.Context) {
	c.JSON(http.StatusOK, &model.FAQResponse{
		APIVersion: handler.apiVersion,
		Timestamp:  time.Now().UnixMilli(),
	})
}
