package controllers

import "github.com/docker/docker/api/types"
import "github.com/docker/docker/client"
import "github.com/humpback/common/models"
import "golang.org/x/net/context"

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// ImageController - handle http request for image
type ImageController struct {
	baseController
}

var imageID string

// Prepare - Format path before exec real action
func (imgCtrl *ImageController) Prepare() {
	imageID = imgCtrl.Ctx.Input.Param(":splat")
}

// GetImage - get image detail with image id or name
func (imgCtrl *ImageController) GetImage() {
	inspectInfo, _, err := dockerClient.ImageInspectWithRaw(context.Background(), imageID)
	if err != nil {
		imgCtrl.Error(500, err.Error())
	}
	imgCtrl.JSON(inspectInfo)
}

// GetImages - get images in this host
func (imgCtrl *ImageController) GetImages() {
	options := types.ImageListOptions{
		All: false,
	}
	images, err := dockerClient.ImageList(context.Background(), options)
	if err != nil {
		imgCtrl.Error(500, err.Error())
	}
	imgCtrl.JSON(images)
}

// PullImage - pull image from respo
func (imgCtrl *ImageController) PullImage() {
	var reqBody models.Image
	if err := json.Unmarshal(imgCtrl.Ctx.Input.RequestBody, &reqBody); err != nil {
		imgCtrl.Error(400, "Invalid json data.")
	}
	if reqBody.Image == "" {
		imgCtrl.Error(400, "Image name cannot be empty or null.")
	}
	err := tryPullImage(reqBody.Image)
	if err != nil {
		imgCtrl.Error(500, err.Error())
	}
}

// DeleteImage - delete image by id or name
func (imgCtrl *ImageController) DeleteImage() {
	var force bool
	force, err := imgCtrl.GetBool("force", false)
	if err != nil {
		force = false
	}

	options := types.ImageRemoveOptions{
		PruneChildren: true,
		Force:         force,
	}
	res, err := dockerClient.ImageRemove(context.Background(), imageID, options)

	if err != nil {
		if strings.Index(err.Error(), "no such image") != -1 {
			imgCtrl.Error(404, err.Error())
		} else {
			imgCtrl.Error(500, err.Error())
		}
	}
	imgCtrl.JSON(res)
}

// tryPullImage - if image is exists then return or pull it from registry
func tryPullImage(imageID string) error {
	_, _, inspectErr := dockerClient.ImageInspectWithRaw(context.Background(), imageID)
	if client.IsErrNotFound(inspectErr) {
		inspectErr = nil
		res, pullErr := dockerClient.ImagePull(context.Background(), imageID, types.ImagePullOptions{})
		if res != nil {
			time.Sleep(time.Second * 3)
			defer res.Close()
		}

		if pullErr != nil {
			return pullErr
		}

		dec := json.NewDecoder(res)
		m := map[string]interface{}{}
		for {
			if err := dec.Decode(&m); err != nil {
				if err == io.EOF {
					break
				}
			}
		}
		// if the final stream object contained an error, return it
		if errMsg, ok := m["error"]; ok {
			return fmt.Errorf("%v", errMsg)
		}
	}
	return inspectErr
}
