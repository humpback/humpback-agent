package model

const (
	//System error codes
	ServerInternalErrorCode = "SYS90000"
	ServerInternalErrorMsg  = "internal server error"
	RequestArgsErrorCode    = "SYS90001"
	RequestArgsErrorMsg     = "request args invalid"
	// Container error codes
	ContainerNotFoundCode            = "CNT10000"
	ContainerNotFoundMsg             = "container not found"
	ContainerCreateErrorCode         = "CNT10001"
	ContainerCreateErrorMsg          = "container create failed"
	ContainerUpdateErrorCode         = "CNT10002"
	ContainerUpdateErrorMsg          = "container update failed"
	ContainerStartErrorCode          = "CNT10003"
	ContainerStartErrorMsg           = "container start failed"
	ContainerCreateConflictErrorCode = "CNT10004"
	ContainerCreateConflictErrorMsg  = "container create conflict"
	// Image error codes
	ImageNotFoundCode  = "IMG10000"
	ImageNotFoundMsg   = "image not found"
	ImagePullErrorCode = "IMG10001"
	ImagePullErrorMsg  = "image pull failed"
)
