package validators

import "github.com/astaxie/beego/context"
import "github.com/humpback/common/models"

import (
	"encoding/json"
	"strings"
)

func CreateServiceValidator(ctx *context.Context) {

	var errs []string
	if ctx.Request.Method == "POST" {
		var reqBody models.CreateProject
		if err := json.Unmarshal(ctx.Input.RequestBody, &reqBody); err != nil {
			errs = append(errs, "Invalid json data.")
		}
		if reqBody.Name == "" {
			errs = append(errs, "Name cannot be empty or null.")
		}
		if reqBody.ComposeData == "" {
			errs = append(errs, "Compose yaml body cannot be empty or null.")
		}
	}

	if ctx.Request.Method == "PUT" {
		var reqBody models.OperateProject
		if err := json.Unmarshal(ctx.Input.RequestBody, &reqBody); err != nil {
			errs = append(errs, "Invalid json data.")
		}
		if reqBody.Name == "" {
			errs = append(errs, "Name cannot be empty or null.")
		}
		if reqBody.Action == "" {
			errs = append(errs, "Action cannot be empty or null.")
		}
		if models.ProjectActions[strings.ToLower(reqBody.Action)] == nil {
			errs = append(errs, "Action type error. We just accept 'start|stop|restart|kill|pause|unpause.")
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
