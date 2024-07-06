package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AmitSuresh/playground/playservices/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

const shutdownTime = 6 * time.Second

func setupServer() (*mux.Router, *zap.Logger) {
	l, _ := zap.NewProduction()
	/* 	wh := handlers.NewWelcomeHandler(l)
	   	rh := handlers.NewReadHandler(l) */
	p := handlers.NewProducts(l)

	//sm := http.NewServeMux()
	sm := mux.NewRouter()

	getRouter := sm.Methods(http.MethodGet).Subrouter()
	getRouter.HandleFunc("/products", p.GetProducts)

	PUTRouter := sm.Methods(http.MethodPut).Subrouter()
	PUTRouter.HandleFunc("/products/{id:[0-9]+}", p.UpdateProduct)
	PUTRouter.Use(p.MiddlewareValidateProduct)

	POSTRouter := sm.Methods(http.MethodPost).Subrouter()
	POSTRouter.HandleFunc("/products", p.AddProduct)
	POSTRouter.Use(p.MiddlewareValidateProduct)

	/*
		sm.Handle("/welcome", wh)
		sm.Handle("/read", rh)
		sm.Handle("/products", p) */
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
