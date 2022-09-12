package webhook

// grpc webhook
import (
	"hboat/grpc/transfer/pool"
	pb "hboat/grpc/transfer/proto"

	"github.com/gin-gonic/gin"
)

var GrpcWebhook = gin.Default()

func init() {
	GrpcWebhook.GET("/api/v1/grpc/all", GrpcGetAgents)
	GrpcWebhook.GET("/api/v1/grpc/config", GrpcSendPlugin)
}

func GrpcGetAgents(c *gin.Context) {
	c.IndentedJSON(200, pool.GlobalGRPCPool.All())
}

func GrpcSendPlugin(c *gin.Context) {
	agentid := c.Query("agentid")
	if agentid == "" {
		c.JSON(500, "agentid is needed")
		return
	}
	connection, err := pool.GlobalGRPCPool.Get(agentid)
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
