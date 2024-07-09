package main

import (
	"flag"
	"fmt"
	"net"

	protos "github.com/AmitSuresh/playground/playservices/v14/currency/v2/protos/currency"
	"github.com/AmitSuresh/playground/playservices/v14/currency/v2/server"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	port = flag.Int("port", 9092, "The server port")
)

func main() {
	flag.Parse()
	log, _ := zap.NewProduction()

	gs := grpc.NewServer()
	csh := server.GetCurrencyServerHandler(log)

	protos.RegisterCurrencyServer(gs, csh)

	reflection.Register(gs)

	// Define the gRPC server options (e.g., port)
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Error("[ERROR]", zap.Any("unable to listen", err))
	}

	// Start the gRPC server
	log.Info("[INFO] Starting gRPC server on port 9092...")
	if err := gs.Serve(listener); err != nil {
		log.Error("[ERROR]", zap.Any("failed to serve", err))
	}
}
