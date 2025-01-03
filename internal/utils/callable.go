package utils

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
)

// 辅助函数: 环境变量转换为 Docker SDK 格式
func MapToEnv(envMap map[string]string) []string {
	var env []string
	for k, v := range envMap {
		env = append(env, k+"="+v)
	}
	return env
}

// 辅助函数：将端口映射转换为 Docker SDK 格式
func MapPorts(portMap map[string]string) nat.PortMap {
	bindings := make(nat.PortMap)
	for hostPort, containerPort := range portMap {
		bindings[nat.Port(containerPort)] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: hostPort,
			},
		}
	}
	return bindings
}

// 辅助函数：将卷映射转换为 Docker SDK 格式
func MapToBinds(volumeMap map[string]string) []string {
	var binds []string
	for hostPath, containerPath := range volumeMap {
		binds = append(binds, hostPath+":"+containerPath)
	}
	return binds
}

// 辅助函数：将设备映射转换为 Docker SDK 格式
func MapToDevices(devices []string) []container.DeviceMapping {
	var deviceMappings []container.DeviceMapping
	for _, device := range devices {
		deviceMappings = append(deviceMappings, container.DeviceMapping{
			PathOnHost:        device,
			PathInContainer:   device,
			CgroupPermissions: "rwm",
		})
	}
	return deviceMappings
}
