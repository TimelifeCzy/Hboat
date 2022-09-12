package pool

import (
	"context"
	"sync"
	"sync/atomic"

	pb "hboat/grpc/transfer/proto"
)

// Connection describe grpc_connection instance by recording
// it's agent and plugin details
//
// From Elkeid
type Connection struct {
	Ctx        context.Context    `json:"-"`
	CancelFunc context.CancelFunc `json:"-"`

	CommandChan chan *pb.Command `json:"-"`

	AgentID    string       `json:"agent_id"`
	Addr       string       `json:"addr"`
	CreateAt   int64        `json:"create_at"`
	LastHBTime atomic.Value `json:"last_heartbeat_time"`

	AgentDetail  map[string]string `json:"agent_detail"`
	agentLock    sync.RWMutex
	PluginDetail map[string]map[string]string `json:"plugin_detail"`
	pluginLock   sync.RWMutex
}

func (c *Connection) GetAgentDetail() map[string]string {
	c.agentLock.RLock()
	defer c.agentLock.RUnlock()
	if c.AgentDetail == nil {
		return map[string]string{}
	}
	return c.AgentDetail
}

func (c *Connection) SetAgentDetail(detail map[string]string) {
	c.agentLock.Lock()
	defer c.agentLock.Unlock()
	c.AgentDetail = detail
}

func (c *Connection) SetPluginDetail(name string, detail map[string]string) {
	c.pluginLock.Lock()
	defer c.pluginLock.Unlock()
	if c.PluginDetail == nil {
		c.PluginDetail = map[string]map[string]string{}
	}
	c.PluginDetail[name] = detail
}

func (c *Connection) GetPluginDetail(name string) map[string]string {
	c.pluginLock.RLock()
	defer c.pluginLock.RUnlock()
	if c.PluginDetail == nil {
		return map[string]string{}
	}
	plgDetail, ok := c.PluginDetail[name]
	if !ok {
		return map[string]string{}
	}
	return plgDetail
}

func (c *Connection) GetPluginsList() []map[string]string {
	c.pluginLock.Lock()
	defer c.pluginLock.Unlock()
	res := make([]map[string]string, 0, len(c.PluginDetail))
	for k := range c.PluginDetail {
		res = append(res, c.PluginDetail[k])
	}
	return res
}
