package main

import (
	"net"

	protos "github.com/AmitSuresh/playground/playservices/v13/protos/currency"
	"github.com/AmitSuresh/playground/playservices/v13/server"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	log, _ := zap.NewProduction()

	gs := grpc.NewServer()
	csh := server.GetCurrencyServerHandler(log)

	protos.RegisterCurrencyServer(gs, csh)

	reflection.Register(gs)

	// Define the gRPC server options (e.g., port)
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Error("[ERROR]", zap.Any("unable to listen", err))
	}

	// Start the gRPC server
	log.Info("[INFO] Starting gRPC server on port 8080...")
	if err := gs.Serve(listener); err != nil {
		log.Error("[ERROR]", zap.Any("failed to serve", err))
	}
}
