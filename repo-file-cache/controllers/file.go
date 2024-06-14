package controllers

import (
	"github.com/opensourceways/repo-file-cache/models"
)

type FileController struct {
	baseController
}

func (fc *FileController) Prepare() {
	fc.apiPrepare()
}

// @Title Set
// @Description store the file
// @Param	body	body 	models.FileUpdateOption		true	"body for storing file"
// @Success 201 {string} "successfully"
// @Failure 400 error_parsing_api_body:     parse payload of request failed
// @Failure 401 error_missing_input_param:  missing some input parameters
// @Failure 402 error_has_same_file:        there are same files which are located in same path
// @Failure 403 error_not_same_file:        the name of files to be stored is not same
// @Failure 500 error_system_error:         system error
// @router / [post]
func (fc *FileController) Set() {
	action := "store the file"

	input := &models.FileUpdateOption{}
	if fr := fc.fetchInputPayload(input); fr != nil {
		fc.sendFailedResultAsResp(fr, action)
		return
	}

	if merr := input.Validate(); merr != nil {
		fc.sendModelErrorAsResp(merr, action)
		return
	}

	if merr := input.Update(); merr != nil {
		fc.sendModelErrorAsResp(merr, action)
		return
	}

	fc.sendResponse(action+" successfully", 0)
}

// @Title Get
// @Description Get the stored file
// @Success 200 {object} models.FilesInfo
// @Failure 400 missing_url_path_parameter: missing url path parameter
// @Failure 500 system_error:               system error
// @router /:filename [get]
func (fc *FileController) Get() {
	action := "list files"

	b, merr := models.ParseBranch(fc.GetString("branch"))
	if merr != nil {
		fc.sendModelErrorAsResp(merr, action)
		return
	}

	summary, _ := fc.GetBool("summary")

	r, merr := models.GetFiles(b, fc.GetString(":filename"), summary)
	if merr != nil {
		fc.sendModelErrorAsResp(merr, action)
		return
	}

	fc.sendSuccessResp(r)
}

// @Title Delete
// @Description Delete the stored files
// @Param	body	body 	models.FileDeleteOption		true	"body for deleting files"
// @Success 204 {string} "successfully"
// @Failure 400 missing_url_path_parameter: missing url path parameter
// @Failure 401 error_parsing_api_body:     parse payload of request failed
// @Failure 402 error_missing_input_param:  missing some input parameters
// @Failure 403 error_not_same_file:        the name of files to be stored is not same
// @Failure 500 system_error:               system error
// @router / [delete]
func (fc *FileController) Delete() {
	action := "delete files"

	input := &models.FileDeleteOption{}
	if fr := fc.fetchInputPayload(input); fr != nil {
		fc.sendFailedResultAsResp(fr, action)
		return
	}

	if merr := input.DeleteFiles(); merr != nil {
		fc.sendModelErrorAsResp(merr, action)
		return
	}

	fc.sendSuccessResp(action + "successfully")
}
