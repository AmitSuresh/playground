package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	protos "github.com/AmitSuresh/playground/playservices/v14/currency"
	"github.com/AmitSuresh/playground/playservices/v14/product-api/data"
	"github.com/AmitSuresh/playground/playservices/v14/product-api/handlers"
	"github.com/go-openapi/runtime/middleware"
	gohandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const shutdownTime = 6 * time.Second

func setupServer() (*mux.Router, *zap.Logger) {
	l, _ := zap.NewProduction()
	v := data.NewValidation()
	conn, err := grpc.Dial("localhost:9092", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	// create client
	cc := protos.NewCurrencyClient(conn)

	// create the handlers
	ph := handlers.NewProducts(l, v, cc)

	//sm := http.NewServeMux()
	sm := mux.NewRouter()

	// handlers for API
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

	// handler for documentation
	opts := middleware.RedocOpts{SpecURL: "/swagger.yaml"}
	sh := middleware.Redoc(opts, nil)

	getR.Handle("/docs", sh)
	getR.Handle("/swagger.yaml", http.FileServer(http.Dir("./")))

	return sm, l
}

func main() {
	sm, l := setupServer()
	// CORS
	ch := gohandlers.CORS(gohandlers.AllowedOrigins([]string{"*"}))
	s := &http.Server{
		Addr:         ":9090",
		Handler:      ch(sm),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 4 * time.Second,
	}
	go func() {
		err := s.ListenAndServe()
		if err != http.ErrServerClosed {
			l.Fatal("error starting server", zap.Any("err", err))
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan
	l.Info("here")
	switch sig {
	case os.Interrupt:
		l.Info("Received interrupt signal")
	default:
		l.Info("Received unknown signal:", zap.Any("signal", sig))
	}

	//http.ListenAndServe("localhost:9090", sm)
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTime)
	defer cancel()

	// Shutdown server with context
	err := s.Shutdown(ctx)
	if err != nil {
		l.Error("error during graceful shutdown", zap.Any("err", err))
	}
}
