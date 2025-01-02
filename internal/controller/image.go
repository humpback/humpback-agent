package controller

import (
	"github.com/docker/docker/client"
	v1model "humpback-agent/internal/api/v1/model"
)

type ImageControllerInterface interface {
	BaseController() ControllerInterface
	Get(request *v1model.GetImageRequest) *v1model.ObjectResult
	List(request *v1model.QueryImageRequest) *v1model.ObjectResult
	Push(request *v1model.PushImageRequest) *v1model.ObjectResult
	Pull(request *v1model.PullImageRequest) *v1model.ObjectResult
	Delete(request *v1model.DeleteImageRequest) *v1model.ObjectResult
	Inspect(request *v1model.InspectImageRequest) *v1model.ObjectResult
}

type ImageController struct {
	baseController ControllerInterface
	client         *client.Client
}

func NewImageController(baseController ControllerInterface, client *client.Client) ImageControllerInterface {
	return &ImageController{
		baseController: baseController,
		client:         client,
	}
}

func (controller *ImageController) BaseController() ControllerInterface {
	return controller.baseController
}

func (controller *ImageController) Get(request *v1model.GetImageRequest) *v1model.ObjectResult {
	return nil
}

func (controller *ImageController) List(request *v1model.QueryImageRequest) *v1model.ObjectResult {
	return nil
}

func (controller *ImageController) Push(request *v1model.PushImageRequest) *v1model.ObjectResult {
	return nil
}

func (controller *ImageController) Pull(request *v1model.PullImageRequest) *v1model.ObjectResult {
	return nil
}

func (controller *ImageController) Delete(request *v1model.DeleteImageRequest) *v1model.ObjectResult {
	return nil
}

func (controller *ImageController) Inspect(request *v1model.InspectImageRequest) *v1model.ObjectResult {
	return nil
}
