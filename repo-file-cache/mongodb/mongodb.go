package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/opensourceways/repo-file-cache/config"
	"github.com/opensourceways/repo-file-cache/dbmodels"
)

var _ dbmodels.IDB = (*client)(nil)

func Initialize(cfg *config.MongodbConfig) (*client, error) {
	c, err := mongo.NewClient(options.Client().ApplyURI(cfg.MongodbConn))
	if err != nil {
		return nil, err
	}

	withContext(func(ctx context.Context) {
		err = c.Connect(ctx)
	})
	if err != nil {
		return nil, err
	}

	// verify if database connection is created successfully
	withContext(func(ctx context.Context) {
		err = c.Ping(ctx, nil)
	})
	if err != nil {
		return nil, err
	}

	cli := &client{
		c:               c,
		db:              c.Database(cfg.DBName),
		filesCollection: cfg.FilesCollection,
	}
	return cli, nil
}

type client struct {
	c  *mongo.Client
	db *mongo.Database

	filesCollection string
}

func (cl *client) Close() error {
	var err error
	withContext(func(ctx context.Context) {
		err = cl.c.Disconnect(ctx)
	})
	return err
}

func (c *client) collection(name string) *mongo.Collection {
	return c.db.Collection(name)
}
