package handler

import (
	"github.com/gin-gonic/gin"
	v1model "humpback-agent/api/v1/model"
	"net/http"
)

func (handler *V1Handler) GetNetworkHandleFunc(c *gin.Context) {
}

func (handler *V1Handler) QueryNetworkHandleFunc(c *gin.Context) {

}

func (handler *V1Handler) CreateNetworkHandleFunc(c *gin.Context) {
	request, err := v1model.BindCreateNetworkRequest(c)
	if err != nil {
		c.JSON(err.StatusCode, err)
		return
	}

	handler.taskChan <- &V1Task{NetworkCreateTask, request}
	c.JSON(http.StatusAccepted, v1model.StdAcceptResult())
}

func (handler *V1Handler) UpdateNetworkHandleFunc(c *gin.Context) {

}

func (handler *V1Handler) DeleteNetworkHandleFunc(c *gin.Context) {
	request, err := v1model.BindDeleteNetworkRequest(c)
	if err != nil {
		c.JSON(err.StatusCode, err)
		return
	}

	handler.taskChan <- &V1Task{NetworkDeleteTask, request}
	c.JSON(http.StatusAccepted, v1model.StdAcceptResult())
}
