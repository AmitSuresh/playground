package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AmitSuresh/playground/playservices/handlers"
	"go.uber.org/zap"
)

const shutdownTime = 6 * time.Second

func setupServer() (*http.ServeMux, *zap.Logger) {
	l, _ := zap.NewProduction()
	wh := handlers.NewWelcomeHandler(l)
	rh := handlers.NewReadHandler(l)

	sm := http.NewServeMux()

	sm.Handle("/welcome", wh)
	sm.Handle("/read", rh)
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
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, syscall.SIGTERM)

	sig := <-sigChan
	l.Info("here")
	switch sig {
	case os.Interrupt:
		l.Info("Received interrupt signal")
	case syscall.SIGTERM:
		l.Info("Received termination signal (SIGTERM)")
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
	} else {
		l.Info("Received interrupt  signal:", zap.Any("ctx", ctx))
	}
}
