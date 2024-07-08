package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	protos "github.com/AmitSuresh/playground/playservices/v14/currency/protos/currency"
	"github.com/AmitSuresh/playground/playservices/v14/product-api/data"
	"github.com/AmitSuresh/playground/playservices/v14/product-api/handlers"
	"github.com/go-openapi/runtime/middleware"
	gohandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	shutdownTime   = 6 * time.Second
	grpcServerAddr = ":8080"
	httpServerAddr = ":9090"
)

func setupGRPCClient(logger *zap.Logger) protos.CurrencyClient {
	// Load client TLS certificates
	cert, err := tls.LoadX509KeyPair("client-cert.pem", "client-key.pem")
	if err != nil {
		logger.Fatal("failed to load client TLS certificates", zap.Error(err))
	}

	// Create a certificate pool from the server CA certificate
	caCert, err := os.ReadFile("ca-cert.pem")
	if err != nil {
		logger.Fatal("failed to read CA certificate", zap.Error(err))
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Dial the gRPC server with transport credentials
	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ServerName:   "localhost", // Server's Common Name (CN)
	})
	conn, err := grpc.Dial("localhost:9092", grpc.WithTransportCredentials(creds))
	if err != nil {
		logger.Fatal("failed to dial gRPC server", zap.Error(err))
	}

	// Return the gRPC client
	return protos.NewCurrencyClient(conn)
}

func setupHTTPServer(logger *zap.Logger, cc protos.CurrencyClient) *http.Server {
	v := data.NewValidation()
	ph := handlers.NewProducts(logger, v, cc)

	sm := mux.NewRouter()

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
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Setup gRPC client
	cc := setupGRPCClient(logger)

	// Setup HTTP server
	httpServer := setupHTTPServer(logger, cc)

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
