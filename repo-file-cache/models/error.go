package models

import (
	"errors"
	"fmt"
	"strings"

	"github.com/opensourceways/repo-file-cache/dbmodels"
)

type IModelError interface {
	Error() string
	ErrCode() ModelErrCode
	IsErrorOf(ModelErrCode) bool
}

type ModelErrCode string

func (e ModelErrCode) toModelError() IModelError {
	return newModelError(e, errors.New(strings.ReplaceAll(string(e), "_", " ")))
}

const (
	ErrSystemError      ModelErrCode = "error_system_error"
	ErrUnknownDBError   ModelErrCode = "error_unknown_db_error"
	ErrNotSameFile      ModelErrCode = "error_not_same_file"
	ErrHasSameFile      ModelErrCode = "error_has_same_file"
	ErrMissingParam     ModelErrCode = "error_missing_input_param"
	ErrInvalidBranchKey ModelErrCode = "error_invalid_branch_key"
)

type modelError struct {
	code ModelErrCode
	err  error
}

func (e modelError) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e modelError) ErrCode() ModelErrCode {
	return e.code
}

func (e modelError) IsErrorOf(code ModelErrCode) bool {
	return e.code == code
}

func newModelError(code ModelErrCode, err error) IModelError {
	return modelError{code: code, err: err}
}

func parseDBError(err dbmodels.IDBError) IModelError {
	if err == nil {
		return nil
	}

	switch err.ErrCode() {
	case dbmodels.ErrSystemError:
		return newModelError(ErrSystemError, err)

	default:
		return newModelError(
			ErrUnknownDBError,
			fmt.Errorf("db code:%s, err:%s", err.ErrCode(), err.Error()),
		)
	}
}
