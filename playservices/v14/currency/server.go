package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/AmitSuresh/playground/playservices/v14/currency/data"
	protos "github.com/AmitSuresh/playground/playservices/v14/currency/protos/currency"
	"github.com/AmitSuresh/playground/playservices/v14/currency/server"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	port     = flag.Int("port", 9092, "The server port")
	grpcAddr string
)

func main() {
	flag.Parse()
	log, _ := zap.NewProduction()
	defer log.Sync()

	err := godotenv.Load()
	if err != nil {
		log.Error("error loading .env file")
	}
	grpcAddr = os.Getenv("GRPC_ADDRESS")

	log.Info("Here are some data: ", zap.Any("grpcAddr: ", grpcAddr), zap.Any("port: ", *port))

	erhandler, err := data.GetExchangeRatesHandler(log)
	if err != nil {
		log.Error("error creating new handler", zap.Error(err))
	}
	gs := grpc.NewServer()
	csh := server.GetCurrencyServerHandler(erhandler, log)

	protos.RegisterCurrencyServer(gs, csh)

	reflection.Register(gs)

	// Define the gRPC server options (e.g., port)
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", grpcAddr, *port))
	if err != nil {
		log.Error("unable to listen", zap.Error(err))
	}

	// Start the gRPC server
	log.Info("Starting gRPC server on port 9092...")
	if err := gs.Serve(listener); err != nil {
		log.Error("failed to serve", zap.Error(err))
	}
}
