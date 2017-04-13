package validators

import (
	"encoding/json"
	"strings"
	"common/models"

	"github.com/astaxie/beego/context"
)

// CreateContainerValidator - validate create container request body
func CreateContainerValidator(ctx *context.Context) {
	var errs []string
	if ctx.Request.Method == "POST" {
		var reqBody models.Container

		if err := json.Unmarshal(ctx.Input.RequestBody, &reqBody); err != nil {
			errs = append(errs, "Invalid json data.")
		}
		if reqBody.Name == "" {
			errs = append(errs, "Name cannot be empty or null.")
		}
		if reqBody.Image == "" {
			errs = append(errs, "Image cannot be empty or null.")
		}
		if reqBody.RestartPolicy != "" && models.RestartPolicyType[reqBody.RestartPolicy] == nil {
			errs = append(errs, "RestartPolicy must be 'no'、'always' or 'on-failure'.")
		}
		if reqBody.RestartPolicy == "on-failure" && reqBody.RestartRetryCount == 0 {
			errs = append(errs, "RestartRetryCount must rather then 0")
		}
		if reqBody.SHMSize < 0 {
			errs = append(errs, "ShmSize cannot less then 0")
		}
		if reqBody.CPUShares < 0 {
			errs = append(errs, "CPUShares cannot less then 0")
		}
		if reqBody.Memory < 0 {
			errs = append(errs, "Memory cannot less then 0")
		}
	}

	if ctx.Request.Method == "PUT" {
		var reqBody models.ContainerOperate
		if err := json.Unmarshal(ctx.Input.RequestBody, &reqBody); err != nil {
			errs = append(errs, "Invalid json data.")
		}
		if reqBody.Container == "" {
			errs = append(errs, "Container cannot be empty or null.")
		}
		if reqBody.Action == "" {
			errs = append(errs, "Action cannot be empty or null.")
		}
		if models.ActionType[strings.ToLower(reqBody.Action)] == nil {
			errs = append(errs, "Action type error. We just accept 'Start|Stop|Restart|Kill|Pause|Unpause.")
		}
		if strings.ToLower(reqBody.Action) == "upgrade" && reqBody.ImageTag == "" {
			errs = append(errs, "ImageTag cannot be empty or null.")
		}
		if strings.ToLower(reqBody.Action) == "rename" && reqBody.NewName == "" {
			errs = append(errs, "NewName cannot be empty or null.")
		}
	}

	if len(errs) > 0 {
		errData := map[string]interface{}{
			"Code":   400,
			"Detail": errs,
		}
		ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
		ctx.ResponseWriter.WriteHeader(400)
		body, _ := json.Marshal(errData)
		ctx.ResponseWriter.Write(body)
	}
}
