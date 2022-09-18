package api

import (
	"fmt"
	gApi "hboat/server/api/grpc"

	"github.com/gin-gonic/gin"
)

func RunGrpcServer(port int) {
	router := gin.Default()
	rGroup := router.Group("/api/v1/grpc/")
	// TODO: auth middleware
	rGroup.POST("/command", gApi.SendCommand)
	rGroup.GET("/conn/idlist", gApi.AgentIDs)
	rGroup.GET("/conn/count", gApi.AgentCount)
	rGroup.GET("/conn/all", gApi.ConnStat)
	rGroup.GET("/conn/stat", gApi.AgentStat)
	rGroup.GET("/conn/basic", gApi.AgentBasic)

	router.Run(fmt.Sprintf(":%d", port))
}
