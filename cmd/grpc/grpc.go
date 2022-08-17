package grpc

import (
	"hboat/cmd/root"
	"hboat/grpc"

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

func init() {
	grpcCommand.PersistentFlags().BoolVar(&enableCA, "ca", true, "enable ca")
	grpcCommand.PersistentFlags().IntVar(&port, "port", 8888, "grpc serve port")
	grpcCommand.PersistentFlags().StringVar(&addr, "addr", "0.0.0.0", "grpc serve address, set to localhost if you need")
	root.RootCommand.AddCommand(grpcCommand)
}

func grpcFunc(command *cobra.Command, args []string) {
	grpc.RunWrapper(enableCA, addr, port)
}
