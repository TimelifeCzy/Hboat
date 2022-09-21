package grpc

import (
	"context"
	"hboat/config"
	ds "hboat/datasource"
	"hboat/grpc/transfer/pool"
	"hboat/server/api/common"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client
var statusCollection *mongo.Collection

type AgentStatus struct {
	AgentID      string                            `bson:"agent_id"`
	Addr         string                            `bson:"addr"`
	CreateAt     int64                             `bson:"create_at"`
	LastHBTime   int64                             `bson:"last_heartbeat_time"`
	AgentDetail  map[string]interface{}            `bson:"agent_detail"`
	PluginDetail map[string]map[string]interface{} `bson:"plugin_detail"`
}

func AgentIDs(c *gin.Context) {
	agentIDs := make([]string, 0)
	for _, v := range pool.GlobalGRPCPool.All() {
		agentIDs = append(agentIDs, v.AgentID)
	}
	common.Response(c, common.SuccessCode, agentIDs)
}

func AgentCount(c *gin.Context) {
	common.Response(c, common.SuccessCode, pool.GlobalGRPCPool.Count())
}

type ConnStatRsp struct {
	AgentInfo   map[string]interface{}            `json:"agent_info"`
	PluginsInfo map[string]map[string]interface{} `json:"plugins_info"`
}

func AgentStat(c *gin.Context) {
	agentid := c.Query("agent_id")
	var as AgentStatus
	err := statusCollection.FindOne(context.Background(), bson.M{"agent_id": agentid}).Decode(&as)
	if err != nil {
		common.Response(c, common.ErrorCode, err.Error())
		return
	}
	res := ConnStatRsp{
		AgentInfo:   as.AgentDetail,
		PluginsInfo: as.PluginDetail,
	}
	common.Response(c, common.SuccessCode, res)
}

type AgentBasicResp struct {
	AgentID  string      `json:"agent_id"`
	Hostname interface{} `json:"hostname"`
	Status   int         `json:"status"`
	CreateAt int64       `json:"create_at"`
	Platform interface{} `json:"platform"`
	Addr     interface{} `json:"addr"`
}

func AgentBasic(c *gin.Context) {
	pageNum := c.GetInt64("pageNum")
	pageSize := c.GetInt64("pageSize")
	skip := (pageNum - 1) * pageSize
	// options
	options := options.Find().SetSort(bson.D{{Key: "create_at", Value: -1}})
	options.Skip = &skip
	options.Limit = &pageSize
	// find
	cur, err := statusCollection.Find(context.Background(), bson.D{})
	if err != nil {
		common.Response(c, common.ErrorCode, err.Error())
		return
	}
	defer cur.Close(context.Background())
	resList := make([]AgentBasicResp, 0, 10)
	now := time.Now().Unix()
	for cur.Next(context.Background()) {
		var as AgentStatus
		if err := cur.Decode(&as); err != nil {
			continue
		}
		detail := as.AgentDetail
		var status int
		d := now - as.LastHBTime
		if d <= 3*60 {
			status = 0
		} else if d >= 30*60 {
			status = 2
		} else {
			status = 1
		}
		tmp := AgentBasicResp{
			AgentID:  as.AgentID,
			Hostname: detail["hostname"],
			Status:   status,
			CreateAt: as.CreateAt,
			Platform: detail["platform"],
			Addr:     as.Addr,
		}
		resList = append(resList, tmp)
	}
	common.Response(c, common.SuccessCode, resList)
}

func init() {
	var err error
	mongoClient, err = ds.NewMongoDB(config.MongoUri, 5)
	if err != nil {
		panic(err)
	}
	statusCollection = mongoClient.Database(ds.Database).Collection(config.MAgentStatusCollection)
}
