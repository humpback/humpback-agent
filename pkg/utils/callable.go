package utils

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"reflect"
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

func RemoveDuplicatesElement[T comparable](s []T) []T {
	result := make([]T, 0)
	m := make(map[T]bool)
	for _, v := range s {
		if _, ok := m[v]; !ok {
			result = append(result, v)
			m[v] = true
		}
	}
	return result
}

func Contains(obj interface{}, target interface{}) bool {
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true
		}
	}
	return false
}

func MergeSlice[T comparable](dest, new []T) []T {
	uniqueMap := make(map[T]struct{})
	for _, v := range dest {
		uniqueMap[v] = struct{}{}
	}

	for _, v := range new {
		if _, exists := uniqueMap[v]; !exists {
			uniqueMap[v] = struct{}{}
			dest = append(dest, v)
		}
	}
	return dest
}

func RemoveFromSlice[T comparable](dest, rem []T) []T {
	removeMap := make(map[T]struct{})
	for _, v := range rem {
		removeMap[v] = struct{}{}
	}

	result := make([]T, 0, len(dest))
	for _, v := range dest {
		if _, exists := removeMap[v]; !exists {
			result = append(result, v)
		}
	}
	return result
}
