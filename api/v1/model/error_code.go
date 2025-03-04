package model

const (
	//System error codes
	ServerInternalErrorCode = "SYS90000"
	ServerInternalErrorMsg  = "internal server error"
	RequestArgsErrorCode    = "SYS90001"
	RequestArgsErrorMsg     = "request args invalid"
	//Container error codes
	ContainerNotFoundCode    = "CNT10000"
	ContainerCreateErrorCode = "CNT10001"
	ContainerDeleteErrorCode = "CNT10002"
	ContainerLogsErrorCode   = "CNT10003"
	ContainerGetErrorCode    = "CNT10004"
	//Image error codes
	ImageNotFoundCode  = "IMG10000"
	ImagePullErrorCode = "IMG10001"
	//Network error codes
	NetworkNotFoundCode    = "NET10000"
	NetworkCreateErrorCode = "NET10001"
	NetworkDeleteErrorCode = "NET10002"
)
