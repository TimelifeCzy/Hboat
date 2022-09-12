package pool

import (
	"errors"
	pb "hboat/grpc/transfer/proto"
	"sync"
	"time"
)

// TODO just testing
const MaxConnection = 1000

var GlobalGRPCPool = NewGRPCPool()

type GRPCPool struct {
	// connPool cache the grpc connections
	// key is agent id and value is *Connection
	connPool map[string]*Connection
	connLock sync.RWMutex
}

func NewGRPCPool() *GRPCPool {
	return &GRPCPool{
		connPool: make(map[string]*Connection),
	}
}

func (g *GRPCPool) Get(agentID string) (*Connection, error) {
	g.connLock.RLock()
	defer g.connLock.RUnlock()
	conn, ok := g.connPool[agentID]
	if !ok {
		return nil, errors.New("agentID not found")
	}
	return conn, nil
}

func (g *GRPCPool) Add(agentID string, conn *Connection) error {
	_, err := g.Get(agentID)
	if err == nil {
		return errors.New("agentID already exists")
	}
	g.connLock.Lock()
	defer g.connLock.Unlock()
	g.connPool[agentID] = conn
	return nil
}

func (g *GRPCPool) Delete(agentID string) {
	g.connLock.Lock()
	defer g.connLock.Unlock()
	delete(g.connPool, agentID)
}

func (g *GRPCPool) Count() int {
	g.connLock.RLock()
	defer g.connLock.RUnlock()
	return len(g.connPool)
}

func (g *GRPCPool) SendCommand(agentID string, command *pb.Command) (err error) {
	conn, err := g.Get(agentID)
	if err != nil {
		return err
	}
	select {
	case conn.CommandChan <- command:
	case <-time.After(3 * time.Second):
		return errors.New("command send timeout 3s")
	}
	// After sending the command, a wating action like Elkied should be implemented
	// for knowning the result of the command execution, use a notify latter
	// TODO
	return nil
}

func (g *GRPCPool) All() []*Connection {
	res := make([]*Connection, 0)
	g.connLock.RLock()
	defer g.connLock.RUnlock()
	for _, v := range g.connPool {
		conn := v
		res = append(res, conn)
	}
	return res
}
