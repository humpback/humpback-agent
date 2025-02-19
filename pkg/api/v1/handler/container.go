package handler

import (
	"github.com/gin-gonic/gin"
	v1model "humpback-agent/pkg/api/v1/model"
	"net/http"
)

func (handler *V1Handler) GetContainerHandleFunc(c *gin.Context) {
	request, err := v1model.BindGetContainerRequest(c)
	if err != nil {
		c.JSON(err.StatusCode, err)
		return
	}

	result := handler.Container().Get(c.Request.Context(), request)
	if result.Error != nil {
		c.JSON(result.Error.StatusCode, result.Error)
		return
	}
	c.JSON(http.StatusOK, result.Object)
}

func (handler *V1Handler) QueryContainerHandleFunc(c *gin.Context) {
	request, err := v1model.BindQueryContainerRequest(c)
	if err != nil {
		c.JSON(err.StatusCode, err)
		return
	}

	result := handler.Container().List(c.Request.Context(), request)
	if result.Error != nil {
		c.JSON(result.Error.StatusCode, result.Error)
		return
	}
	c.JSON(http.StatusOK, result.Object)
}

func (handler *V1Handler) CreateContainerHandleFunc(c *gin.Context) {
	request, err := v1model.BindCreateContainerRequest(c)
	if err != nil {
		c.JSON(err.StatusCode, err)
		return
	}

	handler.taskChan <- &V1Task{ContainerCreateTask, request}
	c.JSON(http.StatusAccepted, v1model.StdAcceptResult())
}

func (handler *V1Handler) UpdateContainerHandleFunc(c *gin.Context) {
}

func (handler *V1Handler) DeleteContainerHandleFunc(c *gin.Context) {
	request, err := v1model.BindDeleteContainerRequest(c)
	if err != nil {
		c.JSON(err.StatusCode, err)
		return
	}

	handler.taskChan <- &V1Task{ContainerDeleteTask, request}
	c.JSON(http.StatusAccepted, v1model.StdAcceptResult())
}

func (handler *V1Handler) RestartContainerHandleFunc(c *gin.Context) {
	request, err := v1model.BindRestartContainerRequest(c)
	if err != nil {
		c.JSON(err.StatusCode, err)
		return
	}

	handler.taskChan <- &V1Task{ContainerRestartTask, request}
	c.JSON(http.StatusAccepted, v1model.StdAcceptResult())
}

func (handler *V1Handler) StartContainerHandleFunc(c *gin.Context) {
	request, err := v1model.BindStartContainerRequest(c)
	if err != nil {
		c.JSON(err.StatusCode, err)
		return
	}

	handler.taskChan <- &V1Task{ContainerStartTask, request}
	c.JSON(http.StatusAccepted, v1model.StdAcceptResult())
}

func (handler *V1Handler) StopContainerHandleFunc(c *gin.Context) {
	request, err := v1model.BindStopContainerRequest(c)
	if err != nil {
		c.JSON(err.StatusCode, err)
		return
	}

	handler.taskChan <- &V1Task{ContainerStopTask, request}
	c.JSON(http.StatusAccepted, v1model.StdAcceptResult())
}
