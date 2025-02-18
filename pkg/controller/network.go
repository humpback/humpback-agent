package controller

import (
	"context"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	v1model "humpback-agent/pkg/api/v1/model"
)

type NetworkControllerInterface interface {
	BaseController() ControllerInterface
	Get(ctx context.Context, request *v1model.GetNetworkRequest) *v1model.ObjectResult
	Create(ctx context.Context, request *v1model.CreateNetworkRequest) *v1model.ObjectResult
}

type NetworkController struct {
	baseController ControllerInterface
	client         *client.Client
}

func NewNetworkController(baseController ControllerInterface, client *client.Client) NetworkControllerInterface {
	return &NetworkController{
		baseController: baseController,
		client:         client,
	}
}

func (controller *NetworkController) BaseController() ControllerInterface {
	return controller.baseController
}

func (controller *NetworkController) Get(ctx context.Context, request *v1model.GetNetworkRequest) *v1model.ObjectResult {
	var networkBody network.Inspect
	if err := controller.baseController.WithTimeout(ctx, func(ctx context.Context) error {
		var networkErr error
		networkBody, networkErr = controller.client.NetworkInspect(ctx, request.NetworkId, network.InspectOptions{})
		return networkErr
	}); err != nil {
		if errdefs.IsNotFound(err) {
			return v1model.ObjectNotFoundErrorResult(v1model.NetworkNotFoundCode, err.Error())
		}
		return v1model.ObjectInternalErrorResult(v1model.ServerInternalErrorCode, err.Error())
	}
	return v1model.ResultWithObject(networkBody)
}

func (controller *NetworkController) Create(ctx context.Context, request *v1model.CreateNetworkRequest) *v1model.ObjectResult {
	ret := controller.Get(ctx, &v1model.GetNetworkRequest{NetworkId: request.NetworkName})
	if ret.Error != nil {
		if ret.Error.Code != v1model.NetworkNotFoundCode {
			return ret
		}
	}

	if ret.Error == nil {
		return v1model.ResultWithObjectId(ret.Object.(network.Inspect).ID)
	}

	var networkInfo network.CreateResponse
	if err := controller.baseController.WithTimeout(ctx, func(ctx context.Context) error {
		var createdErr error
		networkInfo, createdErr = controller.client.NetworkCreate(ctx, request.NetworkName, network.CreateOptions{
			Driver: request.Driver,
			Scope:  "local",
		})
		return createdErr
	}); err != nil {
		return v1model.ObjectInternalErrorResult(v1model.NetworkCreateErrorCode, err.Error())
	}
	return v1model.ResultWithObjectId(networkInfo.ID)
}
