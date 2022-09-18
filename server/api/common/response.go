package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type StatusCode int

const (
	SuccessCode = iota
	ErrorCode
	AuthFailureCode
)

func Response(c *gin.Context, code StatusCode, message interface{}) {
	c.IndentedJSON(
		http.StatusOK, gin.H{
			"code": code,
			"data": message,
		},
	)
}
