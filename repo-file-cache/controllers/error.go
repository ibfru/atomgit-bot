package controllers

import "github.com/opensourceways/repo-file-cache/models"

const (
	errSystemError             = "error_system_error"
	errParsingApiBody          = "error_parsing_api_body"
	errMissingURLPathParameter = "error_missing_url_path_parameter"
)

func parseModelError(err models.IModelError) *failedApiResult {
	if err == nil {
		return nil
	}

	sc := 400
	code := ""
	switch err.ErrCode() {
	case models.ErrUnknownDBError:
		sc = 500
		code = errSystemError

	case models.ErrSystemError:
		sc = 500
		code = errSystemError

	default:
		code = string(err.ErrCode())
	}

	return newFailedApiResult(sc, code, err)
}

func (bc *baseController) sendModelErrorAsResp(err models.IModelError, action string) {
	bc.sendFailedResultAsResp(parseModelError(err), action)
}
