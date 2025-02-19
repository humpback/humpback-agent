package model

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

type GetContainerRequest struct {
	ContainerId string
}

func BindGetContainerRequest(c *gin.Context) (*GetContainerRequest, *ErrorResult) {
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

func BindQueryContainerRequest(c *gin.Context) (*QueryContainerRequest, *ErrorResult) {
	request := QueryContainerRequest{}
	if err := c.Bind(&request); err != nil {
		return nil, RequestErrorResult(RequestArgsErrorCode, RequestArgsErrorMsg)
	}
	return &request, nil
}

type CreateContainerRequest struct {
	ContainerName  string `json:"containerName"`
	*ContainerMeta `json:",inline"`
	*ScheduleInfo  `json:",inline"`
}

func BindCreateContainerRequest(c *gin.Context) (*CreateContainerRequest, *ErrorResult) {
	request := &CreateContainerRequest{}
	if err := c.ShouldBindJSON(request); err != nil {
		return nil, RequestErrorResult(RequestArgsErrorCode, RequestArgsErrorMsg)
	}
	return request, nil
}

//type CreateContainerRequest struct {
//	Name          string               `json:"name"`          // 容器名称
//	Image         string               `json:"image"`         // 镜像名称
//	OnlyCreate    bool                 `json:"onlyCreate"`    // 创建后立即启动
//	AlwaysPull    bool                 `json:"alwaysPull"`    // 是否总是拉取镜像
//	AutoRemove    bool                 `json:"autoRemove"`    // 是否自动删除容器
//	PortMap       map[string]string    `json:"portMap"`       // 端口映射（hostPort:containerPort）
//	PublishAll    bool                 `json:"publishAll"`    // 是否将所有暴露的端口映射到随机主机端口
//	Command       []string             `json:"command"`       // 命令
//	Entrypoint    []string             `json:"entrypoint"`    // 入口点
//	WorkingDir    string               `json:"workingDir"`    // 工作目录
//	Interactive   bool                 `json:"interactive"`   // 是否启用交互模式（-i）
//	TTY           bool                 `json:"tty"`           // 是否启用 TTY（-t）
//	Env           map[string]string    `json:"env"`           // 环境变量
//	Labels        map[string]string    `json:"labels"`        // 标签
//	RestartPolicy string               `json:"restartPolicy"` // 重启策略（never, always, on-failure, unless-stopped）
//	Logger        *ContainerLogger     `json:"logger"`        // Logger 配置
//	Network       *ContainerNetwork    `json:"network"`       // Network 配置
//	Runtime       *ContainerRuntime    `json:"runtime"`       // Runtime 配置
//	Sysctls       *ContainerSysctl     `json:"sysctls"`       // Sysctls 配置
//	Resources     *ContainerResource   `json:"resources"`     // Resource Limits 配置
//	Cap           *ContainerCapability `json:"cap"`           // Capabilities 配置
//}

type UpdateContainerRequest struct{}

type DeleteContainerRequest struct {
	ContainerId string `json:"containerId"`
	Force       bool   `json:"force"`
}

func BindDeleteContainerRequest(c *gin.Context) (*DeleteContainerRequest, *ErrorResult) {
	force := false
	forceQuery := c.Query("force")
	if strings.TrimSpace(forceQuery) != "" {
		value, err := strconv.ParseBool(forceQuery)
		if err != nil {
			return nil, RequestErrorResult(RequestArgsErrorCode, RequestArgsErrorMsg)
		}
		force = value
	}

	containerId := c.Param("containerId")
	return &DeleteContainerRequest{
		ContainerId: containerId,
		Force:       force,
	}, nil
}

type StartContainerRequest struct {
	ContainerId string `json:"containerId"`
}

func BindStartContainerRequest(c *gin.Context) (*StartContainerRequest, *ErrorResult) {
	request := &StartContainerRequest{}
	if err := c.ShouldBind(request); err != nil {
		return nil, RequestErrorResult(RequestArgsErrorCode, RequestArgsErrorMsg)
	}

	if request.ContainerId == "" {
		request.ContainerId = c.Param("containerId")
	}
	return request, nil
}

type StopContainerRequest struct {
	ContainerId string `json:"containerId"`
}

func BindStopContainerRequest(c *gin.Context) (*StopContainerRequest, *ErrorResult) {
	request := &StopContainerRequest{}
	if err := c.ShouldBind(request); err != nil {
		return nil, RequestErrorResult(RequestArgsErrorCode, RequestArgsErrorMsg)
	}

	if request.ContainerId == "" {
		request.ContainerId = c.Param("containerId")
	}
	return request, nil
}

type RestartContainerRequest struct {
	ContainerId string `json:"containerId"`
}

func BindRestartContainerRequest(c *gin.Context) (*RestartContainerRequest, *ErrorResult) {
	request := &RestartContainerRequest{}
	if err := c.ShouldBind(request); err != nil {
		return nil, RequestErrorResult(RequestArgsErrorCode, RequestArgsErrorMsg)
	}

	if request.ContainerId == "" {
		request.ContainerId = c.Param("containerId")
	}
	return request, nil
}
