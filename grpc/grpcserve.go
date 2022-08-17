package grpc

import (
	"hboat/grpc/grpctrans"
	"hboat/grpc/grpctrans/conf"
)

func RunWrapper(enableCA bool, addr string, port int) {
	grpctrans.RunServer(enableCA, addr, port, conf.ServerCert, conf.ServerKey, conf.CaCert)
}
