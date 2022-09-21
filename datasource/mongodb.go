package datasource

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const Database = "hades"

func NewMongoDB(uri string, poolsize uint64) (*mongo.Client, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	var opt options.ClientOptions
	opt.SetMaxPoolSize(poolsize)
	opt.SetReadPreference(readpref.SecondaryPreferred())

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(uri), &opt)
	if err != nil {
		return nil, err
	}

	return mongoClient, nil
}
