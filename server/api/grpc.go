package api

import (
	"fmt"
	gApi "hboat/server/api/grpc"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)

func RunGrpcServer(port int) {
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
	}))
	rGroup := router.Group("/api/v1/grpc/")
	// TODO: auth middleware
	rGroup.POST("/command", gApi.SendCommand)
	rGroup.GET("/conn/count", gApi.AgentCount)
	rGroup.GET("/conn/stat", gApi.AgentStat)
	rGroup.GET("/conn/basic", gApi.AgentBasic)

	router.Run(fmt.Sprintf(":%d", port))
}
