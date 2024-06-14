package mongodb

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/repo-file-cache/dbmodels"
)

const rhForCurrentDir = "-rh_cd-"

func dirToKey(s string) (string, dbmodels.IDBError) {
	if s == dbmodels.FileCurrentDir {
		return fieldFilesItem + dbmodels.FilePathSeparator, nil
	}

	if strings.Contains(s, dbmodels.FileCurrentDir) {
		s = strings.ReplaceAll(s, dbmodels.FileCurrentDir, rhForCurrentDir)
	}

	return fieldFilesItem + s, nil
}

func fullPathOfFile(p, name string) string {
	if p == dbmodels.FilePathSeparator {
		return name
	}
	if strings.Contains(p, rhForCurrentDir) {
		p = strings.ReplaceAll(p, rhForCurrentDir, dbmodels.FileCurrentDir)
	}
	return filepath.Join(p, name)
}

func docFilterOfFiles(b *dbmodels.Branch, fileName string) bson.M {
	branch := fmt.Sprintf("%s:%s:%s:%s", b.Platform, b.Org, b.Repo, b.Branch)
	return bson.M{
		fieldBranch: branch,
		fieldName:   fileName,
	}
}

func toDocOfFiles(branchSHA string, files []dbmodels.File) (bson.M, dbmodels.IDBError) {
	m := make(bson.M, len(files))
	for i := range files {
		item := &files[i]

		key, err := dirToKey(item.Dir())
		if err != nil {
			return nil, err
		}

		m[key] = bson.M{
			fieldKey:   item.SHA,
			fieldValue: item.Content,
		}
	}

	m[fieldSHA] = branchSHA
	return m, nil
}

func (cl *client) UpdateFiles(branch *dbmodels.Branch, branchSHA string, files []dbmodels.File) dbmodels.IDBError {
	if len(files) == 0 {
		return nil
	}

	var err dbmodels.IDBError

	withContext(func(ctx context.Context) {
		docFilter := docFilterOfFiles(branch, files[0].Name())

		var doc bson.M
		doc, err = toDocOfFiles(branchSHA, files)
		if err == nil {
			err = cl.insertDocIfNotExist(ctx, cl.filesCollection, docFilter, doc)
		}
	})

	return err
}

func (cl *client) DeleteFiles(branch *dbmodels.Branch, files []dbmodels.FilePath) dbmodels.IDBError {
	if len(files) == 0 {
		return nil
	}

	fields := make([]string, 0, len(files))
	for _, f := range files {
		if key, err := dirToKey(f.Dir()); err == nil {
			fields = append(fields, key)
		}
	}

	if len(fields) == 0 {
		return nil
	}

	var err dbmodels.IDBError

	withContext(func(ctx context.Context) {
		docFilter := docFilterOfFiles(branch, files[0].Name())
		err = cl.deleteFields(ctx, cl.filesCollection, docFilter, fields)
	})

	return err
}

func (cl *client) GetFiles(branch *dbmodels.Branch, fileName string) (string, []dbmodels.File, dbmodels.IDBError) {
	var v dFiles
	var err dbmodels.IDBError

	withContext(func(ctx context.Context) {
		err = cl.getDoc(
			ctx, cl.filesCollection,
			docFilterOfFiles(branch, fileName),
			bson.M{
				fieldSHA:   1,
				fieldFiles: 1,
			},
			&v,
		)
	})

	if err != nil {
		return "", nil, err
	}

	r := make([]dbmodels.File, 0, len(v.Files))
	for key, info := range v.Files {
		r = append(r, dbmodels.File{
			Path:    dbmodels.FilePath(fullPathOfFile(key, fileName)),
			SHA:     info.SHA,
			Content: info.Content,
		})
	}
	return v.BranchSHA, r, nil
}

func (cl *client) GetFileSummary(branch *dbmodels.Branch, fileName string) ([]dbmodels.File, dbmodels.IDBError) {
	pipeline := bson.A{
		bson.M{"$match": docFilterOfFiles(branch, fileName)},
		bson.M{"$project": bson.M{
			fieldFiles: bson.M{
				"$objectToArray": fieldFilesRef,
			},
		}},
		bson.M{"$project": bson.M{
			fieldFilesVal: 0,
		}},
		bson.M{"$project": bson.M{
			fieldFiles: bson.M{
				"$arrayToObject": fieldFilesRef,
			},
		}},
	}

	var v []dFiles
	f := func(ctx context.Context) error {
		col := cl.collection(cl.filesCollection)
		cursor, err := col.Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}

		return cursor.All(ctx, &v)
	}

	var err dbmodels.IDBError
	withContext(func(ctx context.Context) {
		if e := f(ctx); e != nil {
			err = newSystemError(e)
		}
	})

	if err != nil {
		return nil, err
	}

	if len(v) == 0 {
		return nil, nil
	}

	files := v[0].Files
	r := make([]dbmodels.File, 0, len(files))
	for key, info := range files {
		r = append(r, dbmodels.File{
			Path: dbmodels.FilePath(fullPathOfFile(key, fileName)),
			SHA:  info.SHA,
		})
	}
	return r, nil
}
