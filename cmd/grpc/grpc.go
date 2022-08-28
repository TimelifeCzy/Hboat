package grpc

import (
	"fmt"
	"hboat/cmd/root"
	"hboat/grpc"
	"hboat/server/webhook"

	"github.com/spf13/cobra"
)

var grpcCommand = &cobra.Command{
	Use:   "grpc",
	Short: "hboat grpc server",
	Long:  `Hboat grpc server launcher`,
	Run:   grpcFunc,
}

var enableCA bool
var port int
var addr string

var (
	wport int
	waddr string
)

func init() {
	grpcCommand.PersistentFlags().BoolVar(&enableCA, "ca", false, "enable ca")
	grpcCommand.PersistentFlags().IntVar(&port, "port", 8888, "grpc serve port")
	grpcCommand.PersistentFlags().StringVar(&addr, "addr", "0.0.0.0", "grpc serve address, set to localhost if you need")
	grpcCommand.PersistentFlags().IntVar(&wport, "wport", 7811, "grpc web serve port")
	grpcCommand.PersistentFlags().StringVar(&waddr, "waddr", "0.0.0.0", "grpc serve address, set to localhost if you need")
	root.RootCommand.AddCommand(grpcCommand)
}

func grpcFunc(command *cobra.Command, args []string) {
	go webhook.GrpcWebhook.Run(fmt.Sprintf("%s:%d", waddr, wport))
	grpc.RunWrapper(enableCA, addr, port)
}
