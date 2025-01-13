package controller

import (
	"context"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"humpback-agent/pkg/api/v1/model"
)

type ImageControllerInterface interface {
	BaseController() ControllerInterface
	Get(ctx context.Context, request *model.GetImageRequest) *model.ObjectResult
	List(ctx context.Context, request *model.QueryImageRequest) *model.ObjectResult
	Push(ctx context.Context, request *model.PushImageRequest) *model.ObjectResult
	Pull(ctx context.Context, request *model.PullImageRequest) *model.ObjectResult
	Delete(ctx context.Context, request *model.DeleteImageRequest) *model.ObjectResult
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

func (controller *ImageController) Get(ctx context.Context, request *model.GetImageRequest) *model.ObjectResult {
	return nil
}

func (controller *ImageController) List(ctx context.Context, request *model.QueryImageRequest) *model.ObjectResult {
	return nil
}

func (controller *ImageController) Push(ctx context.Context, request *model.PushImageRequest) *model.ObjectResult {
	return nil
}

func (controller *ImageController) Pull(ctx context.Context, request *model.PullImageRequest) *model.ObjectResult {
	pullOptions := image.PullOptions{
		All:      request.All,
		Platform: request.Platform,
	}

	out, err := controller.client.ImagePull(context.Background(), request.Image, pullOptions)
	if err != nil {
		return model.ObjectNotFoundErrorResult(model.ImagePullErrorCode, model.ImagePullErrorMsg)
	}

	defer out.Close()
	imageInfo, _, err := controller.client.ImageInspectWithRaw(ctx, request.Image)
	if err != nil {
		return model.ObjectNotFoundErrorResult(model.ImagePullErrorCode, model.ImagePullErrorMsg)
	}
	return model.ResultWithObject(imageInfo.ID)
}

func (controller *ImageController) Delete(ctx context.Context, request *model.DeleteImageRequest) *model.ObjectResult {
	return nil
}
