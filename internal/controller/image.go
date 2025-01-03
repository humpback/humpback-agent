package controller

import (
	"context"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	v1model "humpback-agent/internal/api/v1/model"
)

type ImageControllerInterface interface {
	BaseController() ControllerInterface
	Get(ctx context.Context, request *v1model.GetImageRequest) *v1model.ObjectResult
	List(ctx context.Context, request *v1model.QueryImageRequest) *v1model.ObjectResult
	Push(ctx context.Context, request *v1model.PushImageRequest) *v1model.ObjectResult
	Pull(ctx context.Context, request *v1model.PullImageRequest) *v1model.ObjectResult
	Delete(ctx context.Context, request *v1model.DeleteImageRequest) *v1model.ObjectResult
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

func (controller *ImageController) Get(ctx context.Context, request *v1model.GetImageRequest) *v1model.ObjectResult {
	return nil
}

func (controller *ImageController) List(ctx context.Context, request *v1model.QueryImageRequest) *v1model.ObjectResult {
	return nil
}

func (controller *ImageController) Push(ctx context.Context, request *v1model.PushImageRequest) *v1model.ObjectResult {
	return nil
}

func (controller *ImageController) Pull(ctx context.Context, request *v1model.PullImageRequest) *v1model.ObjectResult {
	pullOptions := image.PullOptions{
		All:      request.All,
		Platform: request.Platform,
	}

	out, err := controller.client.ImagePull(context.Background(), request.Image, pullOptions)
	if err != nil {
		return v1model.ObjectNotFoundErrorResult(v1model.ImagePullErrorCode, v1model.ImagePullErrorMsg)
	}

	defer out.Close()
	imageInfo, _, err := controller.client.ImageInspectWithRaw(ctx, request.Image)
	if err != nil {
		return v1model.ObjectNotFoundErrorResult(v1model.ImagePullErrorCode, v1model.ImagePullErrorMsg)
	}
	return v1model.ResultWithObject(imageInfo.ID)
}

func (controller *ImageController) Delete(ctx context.Context, request *v1model.DeleteImageRequest) *v1model.ObjectResult {
	return nil
}
