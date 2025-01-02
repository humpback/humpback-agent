package handler

import (
	"github.com/gin-gonic/gin"
	v1model "humpback-agent/internal/api/v1/model"
	"net/http"
)

func (handler *V1Handler) GetContainerHandleFunc(c *gin.Context) {
	request, err := v1model.ResolveGetContainerRequest(c)
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
	request, err := v1model.ResolveQueryContainerRequest(c)
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
}

func (handler *V1Handler) UpdateContainerHandleFunc(c *gin.Context) {

}

func (handler *V1Handler) DeleteContainerHandleFunc(c *gin.Context) {

}
