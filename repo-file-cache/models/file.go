package models

import (
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/opensourceways/repo-file-cache/dbmodels"
)

type File = dbmodels.File
type Branch = dbmodels.Branch
type FilePath = dbmodels.FilePath

type FilesInfo struct {
	BranchSHA string `json:"branch_sha" required:"true"`
	Files     []File `json:"files" required:"true"`
}

func (f *FilesInfo) validate() IModelError {
	if f.BranchSHA == "" {
		return newMissingParam("branch_sha")
	}

	names := sets.NewString()
	dirs := sets.NewString()

	for i := range f.Files {
		item := &f.Files[i]

		if k, b := item.IsMissingParam(); b {
			return newMissingParam(fmt.Sprintf("files.%d.%s", i, k))
		}

		names.Insert(item.Name())
		dirs.Insert(item.Dir())
	}

	if len(names) != 1 {
		return ErrNotSameFile.toModelError()
	}

	if len(dirs) != len(f.Files) {
		return ErrHasSameFile.toModelError()
	}

	return nil
}

func newMissingParam(k string) IModelError {
	return newModelError(ErrMissingParam, errors.New("missing parameter: "+k))
}

type FileUpdateOption struct {
	Branch
	FilesInfo
}

func (f *FileUpdateOption) Validate() IModelError {
	if k, b := f.Branch.IsMissingParam(); b {
		return newMissingParam(k)
	}

	return f.FilesInfo.validate()
}

func (f *FileUpdateOption) Update() IModelError {
	err := dbmodels.GetDB().UpdateFiles(&f.Branch, f.BranchSHA, f.Files)
	if err == nil {
		return nil
	}

	return parseDBError(err)
}

func GetFiles(b Branch, fileName string, summary bool) (FilesInfo, IModelError) {
	var sha string
	var r []File
	var err dbmodels.IDBError
	if summary {
		r, err = dbmodels.GetDB().GetFileSummary(&b, fileName)
	} else {
		sha, r, err = dbmodels.GetDB().GetFiles(&b, fileName)
	}
	if err == nil {
		return FilesInfo{sha, r}, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return FilesInfo{}, nil
	}

	return FilesInfo{}, parseDBError(err)
}

type FileDeleteOption struct {
	Branch
	Files []FilePath `json:"files" required:"true"`
}

func (f *FileDeleteOption) Validate() IModelError {
	if k, b := f.Branch.IsMissingParam(); b {
		return newMissingParam(k)
	}

	if len(f.Files) == 0 {
		return newMissingParam("files")
	}

	fileName := f.Files[0].Name()

	for _, item := range f.Files {
		if fileName != item.Name() {
			return ErrNotSameFile.toModelError()
		}
	}

	return nil
}

func (f *FileDeleteOption) DeleteFiles() IModelError {
	err := dbmodels.GetDB().DeleteFiles(&f.Branch, f.Files)
	if err == nil || err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return nil
	}

	return parseDBError(err)
}
