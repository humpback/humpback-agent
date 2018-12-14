package filters

import (
	"encoding/json"

	"humpback-agent/config"

	"github.com/astaxie/beego/context"
)

func Authorization(ctx *context.Context) {
	var errs []string
	var authHeaders = ctx.Request.Header["Authorization"]
	var statusCode = 400
	var token = ""

	if len(authHeaders) > 0 {
		token = authHeaders[0]
		if token != config.AuthorizationToken {
			errs = append(errs, "Invalid authorization token.")
			statusCode = 403
		}
	} else {
		errs = append(errs, "Please provide authorization token in header.")
	}

	if len(errs) > 0 {
		errData := map[string]interface{}{
			"Code":   statusCode,
			"Detail": errs,
		}
		ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
		ctx.ResponseWriter.WriteHeader(statusCode)
		body, _ := json.Marshal(errData)
		ctx.ResponseWriter.Write(body)
	}
}
