package controllers

import "context"

type InfoController struct {
	baseController
}

// Prepare - Override baseController
func (info *InfoController) Prepare() {

}

// Get - Return docker info
func (info *InfoController) Get() {

	engineInfo, err := dockerClient.Info(context.Background())
	if err != nil {
		info.Error(500, err.Error())
	}
	info.JSON(engineInfo)
}
