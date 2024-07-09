package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/AmitSuresh/playground/playservices/v14/currency/data"
	protos "github.com/AmitSuresh/playground/playservices/v14/currency/protos/currency"
	"github.com/AmitSuresh/playground/playservices/v14/currency/server"
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

	erhandler, err := data.GetExchangeRatesHandler(log)
	if err != nil {
		log.Error("error creating new handler", zap.Error(err))
	}
	gs := grpc.NewServer()
	csh := server.GetCurrencyServerHandler(erhandler, log)

	protos.RegisterCurrencyServer(gs, csh)

	reflection.Register(gs)

	// Define the gRPC server options (e.g., port)
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Error("unable to listen", zap.Error(err))
	}

	// Start the gRPC server
	log.Debug("Starting gRPC server on port 9092...")
	if err := gs.Serve(listener); err != nil {
		log.Error("failed to serve", zap.Error(err))
	}
}
