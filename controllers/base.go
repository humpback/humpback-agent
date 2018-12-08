package controllers

import "github.com/humpback/common/models"
import "humpback-agent/config"
import "github.com/astaxie/beego"
import "github.com/docker/docker/client"

import (
	"encoding/json"
	"os"
	"strings"
)

// baseController - Provider some common func, like 'Error()' .etc
type baseController struct {
	beego.Controller
}

var dockerClient *client.Client

// Init - init basic info
func Init() {
	var conf = config.GetConfig()
	var err error

	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	dockerClient, err = client.NewClient(conf.DockerEndPoint, conf.DockerAPIVersion, nil, defaultHeaders)

	if err != nil {
		beego.Critical("Cannot connect docker.\n Endpoint: %s\n Detail: %v", conf.DockerEndPoint, err)
		os.Exit(2)
	}
}

// Prepare - Format path before exec real action
func (base *baseController) Prepare() {
}

// Json - Return json data
func (base *baseController) JSON(data interface{}) {
	base.Data["json"] = data
	base.ServeJSON()
}

// Stream - Format stream data to json and return
func (base *baseController) Stream(data string) {
	data = strings.TrimSpace(data)
	data = strings.Replace(data, "\r\n", "", -1)
	data = strings.Replace(data, "}{", "},{", -1)
	res := "[" + data + "]"
	base.Ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
	base.Ctx.ResponseWriter.Write([]byte(res))
	base.StopRun()
}

// Error - Unified processing error
func (base *baseController) Error(status int, msg ...interface{}) {

	errData := map[string]interface{}{
		"Code":   status,
		"Detail": msg[0],
	}
	if len(msg) >= 2 && msg[1] != nil {
		errData["Code"] = msg[1]
		if models.ErrorMap[msg[1].(int)] != "" {
			errData["Message"] = models.ErrorMap[msg[1].(int)]
		}
	}

	base.Ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
	base.Ctx.ResponseWriter.WriteHeader(status)
	body, _ := json.Marshal(errData)
	base.Ctx.ResponseWriter.Write(body)
	base.StopRun()
}
