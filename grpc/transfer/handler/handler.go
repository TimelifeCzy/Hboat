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

	"hboat/config"
	ds "hboat/datasource"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/peer"
)

var mongoClient *mongo.Client
var statusCollection *mongo.Collection

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

	ctx, cancelFunc := context.WithCancel(context.Background())
	conn := pool.Connection{
		AgentID:     agentID,
		Addr:        addr,
		CreateAt:    time.Now().Unix(),
		CommandChan: make(chan *pool.Command),
		Ctx:         ctx,
		CancelFunc:  cancelFunc,
	}
	if err = pool.GlobalGRPCPool.Add(agentID, &conn); err != nil {
		return err
	}

	// 更新数据
	// TODO: channel
	_, err = statusCollection.InsertOne(context.Background(), bson.M{"agent_id": agentID,
		"addr": addr, "create_at": conn.CreateAt})
	if err != nil {
		statusCollection.UpdateOne(context.Background(), bson.M{"agent_id": agentID},
			bson.M{"$set": bson.M{"addr": addr, "create_at": conn.CreateAt}})
	}

	defer pool.GlobalGRPCPool.Delete(agentID)
	go receiveData(stream, &conn)
	go sendData(stream, &conn)
	<-conn.Ctx.Done()
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
// Handle processes
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
			statusCollection.UpdateOne(context.Background(), bson.M{"agent_id": req.AgentID},
				bson.M{"$set": bson.M{"agent_detail": data, "last_heartbeat_time": conn.LastHBTime}})
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
			statusCollection.UpdateOne(context.Background(), bson.M{"agent_id": req.AgentID},
				bson.M{"$set": bson.M{"plugin_detail": bson.M{value.Body.Fields["name"]: data}}})
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

func init() {
	var err error
	mongoClient, err = ds.NewMongoDB(config.MongoUri, 5)
	if err != nil {
		panic(err)
	}
	statusCollection = mongoClient.Database(ds.Database).Collection(config.MAgentStatusCollection)
}
