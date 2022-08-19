package webhook

// grpc webhook
import (
	"hboat/grpc/transfer"
	pb "hboat/grpc/transfer/proto"

	"github.com/gin-gonic/gin"
)

var GrpcWebhook = gin.Default()

func init() {
	GrpcWebhook.GET("/api/v1/grpc/all", GrpcGetAgents)
	GrpcWebhook.GET("/api/v1/grpc/config", GrpcSendPlugin)
}

func GrpcGetAgents(c *gin.Context) {
	c.JSON(200, transfer.GlobalGRPCPool.All())
}

func GrpcSendPlugin(c *gin.Context) {
	agentid := c.Query("agentid")
	if agentid == "" {
		c.JSON(500, "agentid is needed")
		return
	}
	connection, err := transfer.GlobalGRPCPool.Get(agentid)
	if err != nil {
		c.JSON(500, err)
		return
	}
	command := &pb.Command{
		Config: []*pb.ConfigItem{
			{
				Name:        c.Query("name"),
				SHA256:      c.Query("sha256"),
				DownloadURL: []string{c.Query("downloadurl")},
				Version:     c.Query("version"),
			},
		},
	}
	connection.CommandChan <- command
	c.JSON(200, "success")
}

// TEST
// command := &pb.Command{
// 	Config: []*pb.ConfigItem{
// 		{
// 			Name:        "eBPF-driver",
// 			SHA256:      "7577c6e392fad13081b6e62c1e61b1071b9221c50c82ada756c2c65840caaa91",
// 			DownloadURL: []string{"http://127.0.0.1:8000/eBPF-driver"},
// 			Version:     "1.0.0",
// 		},
// 	},
// }
// connection.CommandChan <- command
