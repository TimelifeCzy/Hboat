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
