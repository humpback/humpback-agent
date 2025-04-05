package controller

import (
	"context"
	v1model "humpback-agent/api/v1/model"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
)

type ImageInternalControllerInterface interface {
	AttemptPull(ctx context.Context, imageId string, alwaysPull bool, auth v1model.RegistryAuth) *v1model.ObjectResult
}

type ImageControllerInterface interface {
	ImageInternalControllerInterface
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

func (controller *ImageController) AttemptPull(ctx context.Context, imageId string, alwaysPull bool, auth v1model.RegistryAuth) *v1model.ObjectResult {
	pullImage := alwaysPull
	if !pullImage {
		imageResult := controller.BaseController().Image().Get(ctx, &v1model.GetImageRequest{ImageId: imageId})
		if imageResult.Error != nil {
			if imageResult.Error.Code == v1model.ImageNotFoundCode {
				pullImage = true
			}
		} else {
			pullImage = false //本地已存在
		}
	}

	if pullImage { //拉取镜像（如果需要）
		if ret := controller.BaseController().Image().Pull(ctx, &v1model.PullImageRequest{Image: imageId, ServerAddress: auth.ServerAddress, UserName: auth.RegistryUsername, Password: auth.RegistryPassword}); ret.Error != nil {
			return ret
		}
	}
	return v1model.ResultWithObject(imageId)
}

func (controller *ImageController) Get(ctx context.Context, request *v1model.GetImageRequest) *v1model.ObjectResult {
	imageInfo, _, err := controller.client.ImageInspectWithRaw(ctx, request.ImageId)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return v1model.ObjectNotFoundErrorResult(v1model.ImageNotFoundCode, err.Error())
		}
		return v1model.ObjectInternalErrorResult(v1model.ImagePullErrorCode, err.Error())
	}
	return v1model.ResultWithObject(imageInfo)
}

func (controller *ImageController) List(ctx context.Context, request *v1model.QueryImageRequest) *v1model.ObjectResult {
	return nil
}

func (controller *ImageController) Push(ctx context.Context, request *v1model.PushImageRequest) *v1model.ObjectResult {
	return nil
}

func (controller *ImageController) Pull(ctx context.Context, request *v1model.PullImageRequest) *v1model.ObjectResult {

	if request.UserName != "" && request.Password != "" {

		authConfig := registry.AuthConfig{
			Username:      request.UserName,
			Password:      request.Password,
			ServerAddress: request.ServerAddress,
		}

		_, err := controller.client.RegistryLogin(ctx, authConfig)
		if err != nil {
			return v1model.ObjectNotFoundErrorResult(v1model.ImagePullErrorCode, err.Error())
		}
	}

	pullOptions := image.PullOptions{
		All:      request.All,
		Platform: request.Platform,
	}

	out, err := controller.client.ImagePull(context.Background(), request.Image, pullOptions)
	if err != nil {
		return v1model.ObjectNotFoundErrorResult(v1model.ImagePullErrorCode, err.Error())
	}

	defer out.Close()

	// wait pull image
	for {
		_, err := out.Read(make([]byte, 1024))
		if err != nil {
			break
		}
	}

	imageInfo, _, err := controller.client.ImageInspectWithRaw(ctx, request.Image)
	if err != nil {
		return v1model.ObjectInternalErrorResult(v1model.ImagePullErrorCode, err.Error())
	}
	return v1model.ResultWithObject(imageInfo.ID)
}

func (controller *ImageController) Delete(ctx context.Context, request *v1model.DeleteImageRequest) *v1model.ObjectResult {
	return nil
}
