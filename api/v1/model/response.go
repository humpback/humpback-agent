package model

import (
	"net/http"
)

const (
	MSG_SUCCEED = "succeed"
	MSG_FAILED  = "failed"
)

type FAQResponse struct {
	APIVersion string `json:"apiVersion"`
	Timestamp  int64  `json:"timestamp"`
}

type APIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type ErrorResult struct {
	StatusCode int    `json:"statusCode"`
	Code       string `json:"code"`
	ErrMsg     string `json:"errMsg"`
}

func RequestErrorResult(code string, errMsg string) *ErrorResult {
	return &ErrorResult{
		StatusCode: http.StatusBadRequest,
		Code:       code,
		ErrMsg:     errMsg,
	}
}

func NotFoundErrorResult(code string, errMsg string) *ErrorResult {
	return &ErrorResult{
		StatusCode: http.StatusNotFound,
		Code:       code,
		ErrMsg:     errMsg,
	}
}

func InternalErrorResult(code string, errMsg string) *ErrorResult {
	return &ErrorResult{
		StatusCode: http.StatusInternalServerError,
		Code:       code,
		ErrMsg:     errMsg,
	}
}

type StdResult struct {
	Msg   string       `json:"msg"`
	Error *ErrorResult `json:"error,omitempty"`
}

func StdSucceedResult() *StdResult {
	return &StdResult{
		Msg: "succeed",
	}
}

func StdAcceptResult() *StdResult {
	return &StdResult{
		Msg: "accepted",
	}
}

func StdInternalErrorResult(code string, errMsg string) *StdResult {
	return &StdResult{
		Error: InternalErrorResult(code, errMsg),
	}
}

type ObjectResult struct {
	ObjectId string       `json:"objectId,omitempty"`
	Object   any          `json:"object,omitempty"`
	Msg      string       `json:"msg,omitempty"`
	Error    *ErrorResult `json:"error,omitempty"`
}

func ResultWithObjectId(objectId string) *ObjectResult {
	return &ObjectResult{
		ObjectId: objectId,
	}
}

func ResultWithObject(object any) *ObjectResult {
	return &ObjectResult{
		Object: object,
	}
}

func ResultMessageResult(msg string) *ObjectResult {
	return &ObjectResult{
		Msg: msg,
	}
}

func ObjectRequestErrorResult(code string, errMsg string) *ObjectResult {
	return &ObjectResult{
		Error: RequestErrorResult(code, errMsg),
	}
}

func ObjectNotFoundErrorResult(code string, errMsg string) *ObjectResult {
	return &ObjectResult{
		Error: NotFoundErrorResult(code, errMsg),
	}
}

func ObjectInternalErrorResult(code string, errMsg string) *ObjectResult {
	return &ObjectResult{
		Error: InternalErrorResult(code, errMsg),
	}
}
