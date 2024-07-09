package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log"

	protos "github.com/AmitSuresh/playground/playservices/v14/currency"
	"github.com/AmitSuresh/playground/playservices/v14/product-api/v2/data"
	"github.com/AmitSuresh/playground/playservices/v14/product-api/v2/handlers"
	"github.com/go-openapi/runtime/middleware"
	gohandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	shutdownTime   = 6 * time.Second
	httpServerAddr = "localhost:9090"
)

var serverAddr = flag.String("addr", "localhost:9092", "The server address in the format of host:port")

func setupHTTPServer(logger *zap.Logger, cc protos.CurrencyClient) *http.Server {
	v := data.NewValidation()
	ph := handlers.NewProducts(logger, v, cc)

	sm := mux.NewRouter()
	//github.com/AmitSuresh/playground/playservices/v14/currency/v2
	// Handlers for API endpoints
	getR := sm.Methods(http.MethodGet).Subrouter()
	getR.HandleFunc("/products", ph.ListAll)
	getR.HandleFunc("/products/{id:[0-9]+}", ph.ListSingleProduct)

	putR := sm.Methods(http.MethodPut).Subrouter()
	putR.HandleFunc("/products", ph.Update)
	putR.Use(ph.MiddlewareValidateProduct)

	postR := sm.Methods(http.MethodPost).Subrouter()
	postR.HandleFunc("/products", ph.Create)
	postR.Use(ph.MiddlewareValidateProduct)

	deleteR := sm.Methods(http.MethodDelete).Subrouter()
	deleteR.HandleFunc("/products/{id:[0-9]+}", ph.Delete)

	// Handler for documentation
	opts := middleware.RedocOpts{SpecURL: "/swagger.yaml"}
	sh := middleware.Redoc(opts, nil)
	getR.Handle("/docs", sh)
	getR.Handle("/swagger.yaml", http.FileServer(http.Dir("./")))

	return &http.Server{
		Addr:         httpServerAddr,
		Handler:      gohandlers.CORS(gohandlers.AllowedOrigins([]string{"*"}))(sm),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 4 * time.Second,
	}
}

func main() {
	flag.Parse()
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	logger.Info("[INFO]", zap.Any("serverAddr", *serverAddr))
	conn, err := grpc.NewClient(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	client := protos.NewCurrencyClient(conn)

	// Setup HTTP server
	httpServer := setupHTTPServer(logger, client)

	// Start HTTP server
	go func() {
		logger.Info("Starting HTTP server", zap.String("address", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("error starting HTTP server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan

	logger.Info("Shutdown signal received, shutting down servers...")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTime)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("error during HTTP server shutdown", zap.Error(err))
	}

	logger.Info("HTTP server shutdown complete.")
}
