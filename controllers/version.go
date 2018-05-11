package controllers

import "github.com/docker/docker/api/types"

import (
	"context"
)

type VersionController struct {
	baseController
}

// Prepare - Override baseController
func (versionCtrl *VersionController) Prepare() {

}

// Get - Return docker version
func (versionCtrl *VersionController) Get() {

	clientVersion := dockerClient.ClientVersion()
	serverVersion, err := dockerClient.ServerVersion(context.Background())
	if err != nil {
		versionCtrl.Error(500, err.Error())
	}

	versionInfo := struct {
		ClientVersion string        `json:"ClientVersion"`
		ServerVersion types.Version `json:"ServerVersion"`
	}{
		ClientVersion: clientVersion,
		ServerVersion: serverVersion,
	}
	versionCtrl.JSON(versionInfo)
}
