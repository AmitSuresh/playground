package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AmitSuresh/playground/playservices/v7/data"
	"github.com/AmitSuresh/playground/playservices/v7/handlers"
	"github.com/go-openapi/runtime/middleware"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

const shutdownTime = 6 * time.Second

func setupServer() (*mux.Router, *zap.Logger) {
	l, _ := zap.NewProduction()
	v := data.NewValidation()
	ph := handlers.NewProducts(l, v)

	//sm := http.NewServeMux()
	sm := mux.NewRouter()

	// handlers for API
	getR := sm.Methods(http.MethodGet).Subrouter()
	getR.HandleFunc("/products", ph.ListAll)
	getR.HandleFunc("/products/{id:[0-9]+}", ph.listSingleProduct)

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
	s := &http.Server{
		Addr:         ":9090",
		Handler:      sm,
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
