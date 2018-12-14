package models

// ErrorMap - error map msg
var ErrorMap = map[int]string{
	20001: "Container already created, but cannot start",
	20002: "Get container info failed when upgrade container",
	20003: "Rename container failed when upgrade or update container",
	20004: "Try pull image failed when upgrade container",
	20005: "Stop container failed when upgrade or update container",
	20006: "Cannot create new container when upgrade container",
	20007: "Cannot start new container when upgrade container",
	20008: "Cannot delete old container when upgrade container, but image is upgrade succeed",
	21001: "Try pull image failed when create or update container",
}
