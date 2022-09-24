package grpc

import (
	"context"
	"hboat/config"
	ds "hboat/datasource"
	"hboat/server/api/common"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CountRsp struct {
	Total   int64 `json:"total"`
	Online  int64 `json:"online"`
	Offline int64 `json:"offline"`
}

// AgentCount returns the count of agent status.
//
// An agent is online with 2 conditions. status is on, and heartbeat
// available within 30 mins.
func AgentCount(c *gin.Context) {
	total, err := ds.StatusC.CountDocuments(context.Background(), bson.D{})
	if err != nil {
		common.Response(c, common.ErrorCode, err.Error())
		return
	}
	// within 5 mins, it's available
	hbEndtime := time.Now().Unix() - config.AgentHBSec
	online, err := ds.StatusC.CountDocuments(context.Background(), bson.M{
		"status": true, "last_heartbeat_time": bson.M{"$gt": hbEndtime}})
	if err != nil {
		common.Response(c, common.ErrorCode, err.Error())
		return
	}

	resp := CountRsp{}
	resp.Total = total
	resp.Online = online
	resp.Offline = total - online

	common.Response(c, common.SuccessCode, resp)
}

type ConnStatRsp struct {
	AgentInfo   map[string]interface{}            `json:"agent_info"`
	PluginsInfo map[string]map[string]interface{} `json:"plugins_info"`
}

func AgentStat(c *gin.Context) {
	agentid := c.Query("agent_id")
	var as ds.AgentStatus
	err := ds.StatusC.FindOne(context.Background(), bson.M{"agent_id": agentid}).Decode(&as)
	if err != nil {
		common.Response(c, common.ErrorCode, err.Error())
		return
	}
	// add into agentdetail
	agentInfo := as.AgentDetail
	agentInfo["online"] = as.IsOnline()
	agentInfo["last_heartbeat_time"] = as.LastHBTime
	agentInfo["addr"] = as.Addr

	res := ConnStatRsp{
		AgentInfo:   agentInfo,
		PluginsInfo: as.PluginDetail,
	}
	common.Response(c, common.SuccessCode, res)
}

type AgentBasicResp struct {
	AgentID  string      `json:"agent_id"`
	Hostname interface{} `json:"hostname"`
	Status   bool        `json:"status"`
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
	cur, err := ds.StatusC.Find(context.Background(), bson.D{})
	if err != nil {
		common.Response(c, common.ErrorCode, err.Error())
		return
	}
	defer cur.Close(context.Background())
	resList := make([]AgentBasicResp, 0, 10)
	for cur.Next(context.Background()) {
		var as ds.AgentStatus
		if err := cur.Decode(&as); err != nil {
			continue
		}
		detail := as.AgentDetail
		tmp := AgentBasicResp{
			AgentID:  as.AgentID,
			Hostname: detail["hostname"],
			Status:   as.IsOnline(),
			CreateAt: as.CreateAt,
			Platform: detail["platform"],
			Addr:     as.Addr,
		}
		resList = append(resList, tmp)
	}
	common.Response(c, common.SuccessCode, resList)
}
