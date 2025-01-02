package model

import "github.com/gin-gonic/gin"

type GetContainerRequest struct {
	ContainerId string
}

func ResolveGetContainerRequest(c *gin.Context) (*GetContainerRequest, *ErrorResult) {
	containerId := c.Param("containerId")
	return &GetContainerRequest{
		ContainerId: containerId,
	}, nil
}

/*
e.g QueryContainerRequest

	{
		"size": false,
		"all": true,
		"filters": {
			"status": "running",
			"label": "architecture=aarch64"
		}
	}
*/
type QueryContainerRequest struct {
	Size    bool              `json:"size"`    //是否返回带磁盘大小信息
	All     bool              `json:"all"`     //是否包括已停止的容器
	Latest  bool              `json:"latest"`  //是否只返回最近创建的容器
	Since   string            `json:"since"`   //只返回在指定容器之后创建的容器
	Before  string            `json:"before"`  //只返回在指定容器之前创建的容器
	Limit   int               `json:"limit"`   //返回容器的最大数量
	Filters map[string]string `json:"filters"` //过滤容器的条件(filters.Add("status", "running") or filters.Add("label", "env=prod"))
}

func ResolveQueryContainerRequest(c *gin.Context) (*QueryContainerRequest, *ErrorResult) {
	request := QueryContainerRequest{}
	if err := c.Bind(&request); err != nil {
		return nil, RequestErrorResult(RequestArgsErrorCode, RequestArgsErrorMsg)
	}
	return &request, nil
}

type CreateContainerRequest struct {
	//....
}
type UpdateContainerRequest struct{}
type DeleteContainerRequest struct{}
type StartContainerRequest struct{}
type StopContainerRequest struct{}
type RestartContainerRequest struct{}
