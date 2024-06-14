package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/beego/beego/v2/core/logs"
)

type failedApiResult struct {
	reason     error
	errCode    string
	statusCode int
}

func newFailedApiResult(statusCode int, errCode string, err error) *failedApiResult {
	return &failedApiResult{
		statusCode: statusCode,
		errCode:    errCode,
		reason:     err,
	}
}

func (bc *baseController) sendResponse(body interface{}, statusCode int) {
	if statusCode != 0 {
		// if success, don't set status code, otherwise the header set in bc.ServeJSON
		// will not work. The reason maybe the same as above.
		bc.Ctx.ResponseWriter.WriteHeader(statusCode)
	}

	bc.Data["json"] = struct {
		Data interface{} `json:"data"`
	}{
		Data: body,
	}

	bc.ServeJSON()
}

func (bc *baseController) sendSuccessResp(body interface{}) {
	bc.sendResponse(body, 0)
}

func (bc *baseController) sendFailedResponse(statusCode int, errCode string, reason error, action string) {
	if statusCode >= 500 {
		logs.Error(fmt.Sprintf("Failed to %s, errCode: %s, err: %s", action, errCode, reason.Error()))

		errCode = errSystemError
		reason = fmt.Errorf("system error")
	}

	d := struct {
		ErrCode string `json:"error_code"`
		ErrMsg  string `json:"error_message"`
	}{
		ErrCode: fmt.Sprintf("rfc.%s", errCode),
		ErrMsg:  reason.Error(),
	}

	bc.sendResponse(d, statusCode)
}

func (bc *baseController) sendFailedResultAsResp(fr *failedApiResult, action string) {
	bc.sendFailedResponse(fr.statusCode, fr.errCode, fr.reason, action)
}

func (bc *baseController) newFuncForSendingFailedResp(action string) func(fr *failedApiResult) {
	return func(fr *failedApiResult) {
		bc.sendFailedResponse(fr.statusCode, fr.errCode, fr.reason, action)
	}
}

func (bc *baseController) fetchInputPayload(info interface{}) *failedApiResult {
	return fetchInputPayloadData(bc.Ctx.Input.RequestBody, info)
}

func fetchInputPayloadData(input []byte, info interface{}) *failedApiResult {
	if err := json.Unmarshal(input, info); err != nil {
		return newFailedApiResult(
			400, errParsingApiBody, fmt.Errorf("invalid input payload: %s", err.Error()),
		)
	}
	return nil
}

func (bc *baseController) checkPathParameter() *failedApiResult {
	rp := bc.routerPattern()
	if rp == "" {
		return nil
	}

	items := strings.Split(rp, "/")
	for _, item := range items {
		if strings.HasPrefix(item, ":") && bc.GetString(item) == "" {
			return newFailedApiResult(
				400, errMissingURLPathParameter,
				fmt.Errorf("missing path parameter:%s", item),
			)
		}
	}

	return nil
}

func (bc *baseController) routerPattern() string {
	if v, ok := bc.Data["RouterPattern"]; ok {
		return v.(string)
	}
	return ""
}

func (bc *baseController) apiReqHeader(h string) string {
	return bc.Ctx.Input.Header(h)
}

func (bc *baseController) apiRequestMethod() string {
	return bc.Ctx.Request.Method
}

func (bc *baseController) isPostRequest() bool {
	return bc.apiRequestMethod() == http.MethodPost
}
