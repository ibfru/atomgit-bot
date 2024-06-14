package mongodb

import (
	"errors"

	"github.com/opensourceways/repo-file-cache/dbmodels"
)

var (
	errNoDBRecord = dbmodels.NewDBError(dbmodels.ErrNoDBRecord, errors.New("no record"))
)

func newSystemError(err error) dbmodels.IDBError {
	return dbmodels.NewDBError(dbmodels.ErrSystemError, err)
}
