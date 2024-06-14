package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/opensourceways/repo-file-cache/dbmodels"
)

func isErrNoDocuments(err error) bool {
	return err.Error() == mongo.ErrNoDocuments.Error()
}

func withContext(f func(context.Context)) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	f(ctx)
}

func (cl *client) insertDocIfNotExist(ctx context.Context, collection string, filterOfDoc, docInfo bson.M) dbmodels.IDBError {
	upsert := true

	r := cl.collection(collection).FindOneAndUpdate(
		ctx, filterOfDoc, bson.M{"$set": docInfo},
		&options.FindOneAndUpdateOptions{
			Upsert:     &upsert,
			Projection: bson.M{fieldName: 1}, // avoid returning all the fields of the document
		},
	)

	if err := r.Err(); err != nil && !isErrNoDocuments(err) {
		return newSystemError(err)
	}
	return nil
}

func (cl *client) getDoc(ctx context.Context, collection string, filterOfDoc, project bson.M, result interface{}) dbmodels.IDBError {
	col := cl.collection(collection)

	var sr *mongo.SingleResult
	if len(project) > 0 {
		sr = col.FindOne(ctx, filterOfDoc, &options.FindOneOptions{
			Projection: project,
		})
	} else {
		sr = col.FindOne(ctx, filterOfDoc)
	}

	if err := sr.Decode(result); err != nil {
		if isErrNoDocuments(err) {
			return errNoDBRecord
		}
		return newSystemError(err)
	}
	return nil
}

func (cl *client) deleteFields(ctx context.Context, collection string, filterOfDoc bson.M, fields []string) dbmodels.IDBError {
	items := bson.M{}
	for _, item := range fields {
		items[item] = ""
	}

	r, err := cl.collection(collection).UpdateOne(
		ctx, filterOfDoc, bson.M{"$unset": items},
	)
	if err != nil {
		return newSystemError(err)
	}

	if r.MatchedCount == 0 {
		return errNoDBRecord
	}
	return nil
}
