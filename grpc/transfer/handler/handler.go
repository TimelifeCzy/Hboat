package handler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"hboat/grpc/transfer/pool"
	pb "hboat/grpc/transfer/proto"

	"google.golang.org/grpc/peer"
)

// TransferHandler implements svc.TransferServer
type TransferHandler struct{}

func (h *TransferHandler) Transfer(stream pb.Transfer_TransferServer) error {
	data, err := stream.Recv()
	if err != nil {
		return err
	}
	agentID := data.AgentID
	p, ok := peer.FromContext(stream.Context())
	if !ok {
		return errors.New("client ip get error")
	}
	addr := p.Addr.String()
	fmt.Printf("Get connection %s from %s\n", agentID, addr)

	createAt := time.Now().UnixNano() / (1000 * 1000 * 1000)
	ctx, cancelFunc := context.WithCancel(context.Background())
	connection := pool.Connection{
		AgentID:     agentID,
		Addr:        addr,
		CreateAt:    createAt,
		CommandChan: make(chan *pb.Command),
		Ctx:         ctx,
		CancelFunc:  cancelFunc,
	}

	err = pool.GlobalGRPCPool.Add(agentID, &connection)
	if err != nil {
		return err
	}

	defer pool.GlobalGRPCPool.Delete(agentID)
	go receiveData(stream, &connection)
	go sendData(stream, &connection)

	<-connection.Ctx.Done()
	return nil
}

func sendData(stream pb.Transfer_TransferServer, conn *pool.Connection) {
	defer conn.CancelFunc()

	for {
		select {
		case <-conn.Ctx.Done():
			return
		case cmd := <-conn.CommandChan:
			if cmd == nil {
				return
			}
			err := stream.Send(cmd)
			if err != nil {
				return
			}
		}
	}
}

func receiveData(stream pb.Transfer_TransferServer, conn *pool.Connection) {
	defer conn.CancelFunc()
	for {
		select {
		case <-conn.Ctx.Done():
			return
		default:
			data, err := stream.Recv()
			if err != nil {
				return
			}
			handleData(data, conn)
		}
	}
}

func handleData(req *pb.RawData, conn *pool.Connection) {
	interIpv4 := strings.Join(req.IntranetIPv4, ",")
	interIpv6 := strings.Join(req.IntranetIPv6, ",")
	extraIpv4 := strings.Join(req.ExtranetIPv4, ",")
	extraIpv6 := strings.Join(req.ExtranetIPv6, ",")

	for _, v := range req.GetData() {
		dataType := v.DataType
		switch {
		// agent-heartbeat
		case dataType == 1:
			conn.LastHBTime.Store(time.Now().Unix())
			v.Body.Fields["intranet_ipv4"] = interIpv4
			v.Body.Fields["intranet_ipv6"] = interIpv6
			v.Body.Fields["extranet_ipv4"] = extraIpv4
			v.Body.Fields["extranet_ipv6"] = extraIpv6
			v.Body.Fields["product"] = req.Product
			v.Body.Fields["hostname"] = req.Hostname
			v.Body.Fields["version"] = req.Version
			conn.SetAgentDetail(v.Body.Fields)
			fmt.Println("Agent-Heartbeat:", conn.GetAgentDetail())
		// plugin-heartbeat
		case dataType == 2:
			conn.SetPluginDetail(v.Body.Fields["name"], v.Body.Fields)
			fmt.Println("Plugin-HeartBeat:", conn.GetPluginDetail(v.Body.Fields["name"]))
		// windows
		case dataType >= 100 && dataType <= 400:
			for _, item := range req.Item {
				// backport for windows for temp
				ParseWinDataDispatch(item.Fields, req, int(dataType))
			}
		default:
		}
	}
}
