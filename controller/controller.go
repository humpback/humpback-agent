package controller

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	v1model "humpback-agent/api/v1/model"
	"humpback-agent/model"
	"humpback-agent/pkg/utils"
	"path/filepath"
	"regexp"
	"time"
)

// 编译正则表达式
var re = regexp.MustCompile(`\{([^}]+)\}`)

type GetConfigValueFunc func(configNames []string) (map[string][]byte, error)

type InternalController interface {
	WithTimeout(ctx context.Context, callback func(context.Context) error) error
	DockerEngine(ctx context.Context) (*model.DockerEngineInfo, error)
	GetConfigNamesWithVolumes(volumes []*v1model.ServiceVolume) map[string]string
	BuildVolumesWithConfigNames(configNames map[string]string) (map[string]string, error)
	ConfigValues(ctx context.Context, configNames []string) (map[string][]byte, error)
}

type ControllerInterface interface {
	InternalController
	Image() ImageControllerInterface
	Container() ContainerControllerInterface
	Network() NetworkControllerInterface
}

type BaseController struct {
	client               *client.Client
	volumesRootDirectory string
	reqTimeout           time.Duration
	getConfigFunc        GetConfigValueFunc
	image                ImageControllerInterface
	container            ContainerControllerInterface
	network              NetworkControllerInterface
}

func NewController(client *client.Client, getConfigFunc GetConfigValueFunc, volumesRootDirectory string, reqTimeout time.Duration) ControllerInterface {
	baseController := &BaseController{
		client:               client,
		volumesRootDirectory: volumesRootDirectory,
		reqTimeout:           reqTimeout,
		getConfigFunc:        getConfigFunc,
	}

	baseController.image = NewImageController(baseController, client)
	baseController.container = NewContainerController(baseController, client)
	baseController.network = NewNetworkController(baseController, client)
	return baseController
}

func (controller *BaseController) WithTimeout(ctx context.Context, callback func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, controller.reqTimeout)
	defer cancel()
	return callback(ctx)
}

func (controller *BaseController) DockerEngine(ctx context.Context) (*model.DockerEngineInfo, error) {
	var engineInfo model.DockerEngineInfo
	serverVersion, getErr := controller.client.ServerVersion(ctx)
	if getErr != nil {
		return nil, getErr
	}

	dockerInfo, infoErr := controller.client.Info(ctx)
	if infoErr != nil {
		return nil, infoErr
	}

	engineInfo.Version = dockerInfo.ServerVersion
	engineInfo.APIVersion = serverVersion.APIVersion
	engineInfo.RootDirectory = dockerInfo.DockerRootDir
	engineInfo.StorageDriver = dockerInfo.Driver
	engineInfo.LoggingDriver = dockerInfo.LoggingDriver
	engineInfo.VolumePlugins = dockerInfo.Plugins.Volume
	engineInfo.NetworkPlugins = dockerInfo.Plugins.Network
	return &engineInfo, nil
}

func (controller *BaseController) GetConfigNamesWithVolumes(volumes []*v1model.ServiceVolume) map[string]string {
	configPair := map[string]string{}
	for _, volume := range volumes {
		if volume.Type == v1model.ServiceVolumeTypeBind { //先只实现bind类型
			matches := re.FindStringSubmatch(volume.Source)
			if len(matches) > 1 {
				if fileName := filepath.Base(volume.Target); fileName != "" {
					configPair[matches[1]] = fileName
				}
			}
		}
	}
	return configPair
}

func (controller *BaseController) ConfigValues(ctx context.Context, configNames []string) (map[string][]byte, error) {
	if controller.getConfigFunc != nil {
		return controller.getConfigFunc(configNames)
	}
	return nil, fmt.Errorf("no setting config value getter")
}

func (controller *BaseController) BuildVolumesWithConfigNames(configNames map[string]string) (map[string]string, error) {
	configPaths := map[string]string{}
	if len(configNames) > 0 {
		names := []string{}
		for name, _ := range configNames {
			names = append(names, name)
		}
		
		configValues, err := controller.ConfigValues(context.Background(), names)
		if err != nil {
			return nil, err
		}

		var vid uuid.UUID
		for configName, data := range configValues {
			vid, err = uuid.NewUUID()
			if err != nil {
				return nil, err
			}
			if fileName, ret := configNames[configName]; ret {
				filePath := fmt.Sprintf("%s/%s/_data/%s", controller.volumesRootDirectory, vid.String(), fileName)
				if err = utils.WriteFileWithDir(filePath, []byte(data), 0755); err != nil {
					return nil, err
				}
				configPaths[configName] = filePath
			}
		}
	}
	return configPaths, nil
}

func (controller *BaseController) Image() ImageControllerInterface {
	return controller.image
}

func (controller *BaseController) Container() ContainerControllerInterface {
	return controller.container
}

func (controller *BaseController) Network() NetworkControllerInterface {
	return controller.network
}
