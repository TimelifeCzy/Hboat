package datasource

import (
	"context"
	"hboat/config"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	Database       = "hades"
	AgentStatusCol = "agentstatus"
)

var MongoInst *mongo.Client

type AgentStatus struct {
	AgentID      string                            `bson:"agent_id"`
	Addr         string                            `bson:"addr"`
	CreateAt     int64                             `bson:"create_at"`
	LastHBTime   int64                             `bson:"last_heartbeat_time"`
	AgentDetail  map[string]interface{}            `bson:"agent_detail"`
	PluginDetail map[string]map[string]interface{} `bson:"plugin_detail"`
}

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

func init() {
	var err error
	MongoInst, err = NewMongoDB(config.MongoURI, 5)
	if err != nil {
		panic(err)
	}
}
