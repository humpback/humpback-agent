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
	//Image error codes
	ImagePullErrorCode = "IMG10001"
	//Network error codes
	NetworkNotFoundCode    = "NET10000"
	NetworkCreateErrorCode = "NET10001"
)
