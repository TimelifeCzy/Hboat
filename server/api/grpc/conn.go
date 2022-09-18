package grpc

import (
	"hboat/grpc/transfer/pool"
	"hboat/server/api/common"

	"github.com/gin-gonic/gin"
)

func AgentIDs(c *gin.Context) {
	agentIDs := make([]string, 128)
	for _, v := range pool.GlobalGRPCPool.All() {
		agentIDs = append(agentIDs, v.AgentID)
	}
	common.Response(c, common.SuccessCode, agentIDs)
}

func AgentCount(c *gin.Context) {
	common.Response(c, common.SuccessCode, pool.GlobalGRPCPool.Count())
}

type ConnStatRsp struct {
	AgentInfo   map[string]interface{}   `json:"agent_info"`
	PluginsInfo []map[string]interface{} `json:"plugins_info"`
}

func ConnStat(c *gin.Context) {
	res := pool.GlobalGRPCPool.All()
	resList := make([]ConnStatRsp, 0, len(res))
	for _, v := range res {
		tmp := ConnStatRsp{
			AgentInfo:   v.GetAgentDetail(),
			PluginsInfo: v.GetPluginsList(),
		}
		resList = append(resList, tmp)
	}
	common.Response(c, common.SuccessCode, resList)
}

func AgentStat(c *gin.Context) {
	agentid := c.Query("agent_id")
	conn, err := pool.GlobalGRPCPool.Get(agentid)
	if err != nil {
		common.Response(c, common.ErrorCode, err.Error())
		return
	}
	res := ConnStatRsp{
		AgentInfo:   conn.GetAgentDetail(),
		PluginsInfo: conn.GetPluginsList(),
	}
	common.Response(c, common.SuccessCode, res)
}

type AgentBasicResp struct {
	AgentID string `json:"agent_id"`
	Hostname interface{} `json:"hostname"`
	LastHBTime int64 `json:"last_heartbeat_time"`
	CreateAt int64 `json:"create_at"`
	Platform interface{} `json:"platform"`
	Addr interface{} `json:"addr"`
}

func AgentBasic(c *gin.Context) {
	res := pool.GlobalGRPCPool.All()
	resList := make([]AgentBasicResp, 0, len(res))
	for _, v := range res {
		detail := v.GetAgentDetail()		
		tmp := AgentBasicResp{
			AgentID: v.AgentID,
			Hostname: detail["hostname"],
			LastHBTime: v.LastHBTime,
			CreateAt: v.CreateAt,
			Platform: detail["platform"],
			Addr: v.Addr,
		}
		resList = append(resList, tmp)
	}
	common.Response(c, common.SuccessCode, resList)
}
