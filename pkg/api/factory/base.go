package factory

import (
	"errors"
	"github.com/gin-gonic/gin"
	"humpback-agent/pkg/controller"
	"strings"
)

var (
	ErrRegisterHandlerVersionInvalid      = errors.New("register handler version invalid")
	ErrRegisterHandlerVersionAlreadyExist = errors.New("register handler version already registered")
	ErrConstructHandlerNotImplemented     = errors.New("construct handler not implemented")
)

type Initializer interface {
	SetRouter(version string, engine *gin.Engine)
}

type BaseHandler struct {
	Initializer
	controller.ControllerInterface
}

type HandlerInterface any

type ConstructHandlerFunc func(controller controller.ControllerInterface) HandlerInterface

var handlers = map[string]ConstructHandlerFunc{}

func InjectHandler(version string, constructFunc ConstructHandlerFunc) error {
	version = strings.ToLower(strings.TrimSpace(version))
	if version == "" {
		return ErrRegisterHandlerVersionInvalid
	}

	if _, ret := handlers[version]; ret {
		return ErrRegisterHandlerVersionAlreadyExist
	}

	handlers[version] = constructFunc
	return nil
}

func HandlerConstruct(version string, controller controller.ControllerInterface) (any, error) {
	version = strings.ToLower(strings.TrimSpace(version))
	constructFunc, ret := handlers[version]
	if !ret {
		return nil, ErrConstructHandlerNotImplemented
	}

	handler := constructFunc(controller)
	return handler, nil
}
