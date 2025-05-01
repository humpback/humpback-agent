package model

import (
	"github.com/gin-gonic/gin"
)

type GetNetworkRequest struct {
	NetworkId string `json:"networkId"`
}

type CreateNetworkRequest struct {
	NetworkName string `json:"networkName"`
	Driver      string `json:"driver"`
	Scope       string `json:"scope"`
}

func BindCreateNetworkRequest(c *gin.Context) (*CreateNetworkRequest, *ErrorResult) {
	request := &CreateNetworkRequest{}
	if err := c.ShouldBindJSON(request); err != nil {
		return nil, RequestErrorResult(RequestArgsErrorCode, RequestArgsErrorMsg)
	}
	return request, nil
}

type DeleteNetworkRequest struct {
	NetworkId string `json:"networkId"`
	Scope     string `json:"scope"`
}

func BindDeleteNetworkRequest(c *gin.Context) (*DeleteNetworkRequest, *ErrorResult) {
	networkId := c.Param("networkId")
	scope := c.Query("scope")
	return &DeleteNetworkRequest{
		NetworkId: networkId,
		Scope:     scope,
	}, nil
}
