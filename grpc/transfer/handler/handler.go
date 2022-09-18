package handler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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
		CommandChan: make(chan *pool.Command),
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
			err := stream.Send(cmd.Command)
			if err != nil {
				cmd.Error = err
				close(cmd.Ready)
				return
			}
			cmd.Error = nil
			close(cmd.Ready)
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

// handleData handles received data
//
// TODO: heartbeat to influxdb or ES
func handleData(req *pb.RawData, conn *pool.Connection) {
	intranet_ipv4 := strings.Join(req.IntranetIPv4, ",")
	intranet_ipv6 := strings.Join(req.IntranetIPv6, ",")
	extranet_ipv4 := strings.Join(req.ExtranetIPv4, ",")
	extranet_ipv6 := strings.Join(req.ExtranetIPv6, ",")

	for _, value := range req.GetData() {
		dataType := value.DataType
		switch {
		// agent-heartbeat
		case dataType == 1:
			data := make(map[string]interface{}, 40)
			data["intranet_ipv4"] = intranet_ipv4
			data["intranet_ipv6"] = intranet_ipv6
			data["extranet_ipv4"] = extranet_ipv4
			data["extranet_ipv6"] = extranet_ipv6
			data["product"] = req.Product
			data["hostname"] = req.Hostname
			data["version"] = req.Version
			for k, v := range value.Body.Fields {
				// skip special field, hard-code
				if k == "platform_version" || k == "version" {
					data[k] = v
					continue
				}
				fv, err := strconv.ParseFloat(v, 64)
				if err == nil {
					data[k] = fv
				} else {
					data[k] = v
				}
			}
			conn.LastHBTime = time.Now().Unix()
			conn.SetAgentDetail(data)
		// plugin-heartbeat
		case dataType == 2:
			data := make(map[string]interface{}, 20)
			for k, v := range value.Body.Fields {
				// skip special field, hard-code
				if k == "pversion" {
					data[k] = v
					continue
				}
				fv, err := strconv.ParseFloat(v, 64)
				if err == nil {
					data[k] = fv
				} else {
					data[k] = v
				}
			}
			conn.SetPluginDetail(value.Body.Fields["name"], data)
		// windows
		case dataType >= 100 && dataType <= 400:
			for _, item := range req.Item {
				// backport for windows for temp
				ParseWinDataDispatch(item.Fields, req, int(dataType))
			}
		default:
			// TODO
		}
	}
}
