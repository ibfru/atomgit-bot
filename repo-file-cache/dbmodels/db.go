package dbmodels

var db IDB

func RegisterDB(idb IDB) {
	db = idb
}

func GetDB() IDB {
	return db
}

type IDB interface {
	Close() error
	UpdateFiles(branch *Branch, branchSHA string, files []File) IDBError
	DeleteFiles(branch *Branch, files []FilePath) IDBError
	GetFiles(branch *Branch, fileName string) (string, []File, IDBError)
	GetFileSummary(branch *Branch, fileName string) ([]File, IDBError)
}
