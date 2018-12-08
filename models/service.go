package models

import "github.com/docker/docker/api/types"

//ServiceState - define compose service state struct
type ServiceState struct {
	Name string `json:"Name"`
	*types.ContainerState
}
