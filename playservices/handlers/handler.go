package handlers

import (
	"context"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type LoggerKeyType string

var logKey LoggerKeyType = "LOGGER"

type DataKeyType string

var dataKey DataKeyType = "DATA"

type welcomeHandler struct {
	l *zap.Logger
}

func NewWelcomeHandler(l *zap.Logger) *welcomeHandler {
	return &welcomeHandler{l}
}

func (h *welcomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	//log, _ := zap.NewProduction()
	log := h.l
	ctx = context.WithValue(ctx, logKey, log)
	log.Info("Hello World!", zap.Any(string(logKey), ctx.Value(logKey)))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello World!"))
}

type readHandler struct {
	l *zap.Logger
}

func NewReadHandler(l *zap.Logger) *readHandler {
	return &readHandler{l}
}

/*
	 func (h *readHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		//ctx := r.Context()
		//log, _ := zap.NewProduction()
		log := h.l
		ctx = context.WithValue(ctx, logKey, log)
		log.Info("Hello World!", zap.Any(string(logKey), ctx.Value(logKey)))
		d, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatal("error reading", zap.Any("err", err), zap.Any("status code", http.StatusBadRequest))
			http.Error(w, "oops", http.StatusBadRequest)
			return
		}
		ctx = context.WithValue(ctx, dataKey, d)
		log.Info("reading!", zap.Any(string(dataKey), ctx.Value(dataKey)))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("reading!"))
	}
*/
func (h *readHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel() // Ensure cancellation on exit

	log := h.l
	ctx = context.WithValue(ctx, logKey, log)
	log.Info("Hello World!", zap.Any(string(logKey), ctx.Value(logKey)))

	d, err := io.ReadAll(r.Body)
	defer r.Body.Close() // Close body regardless of error
	if err != nil {
		log.Fatal("error reading", zap.Any("err", err), zap.Any("status code", http.StatusBadRequest))
		http.Error(w, "oops", http.StatusBadRequest)
		return
	}

	ctx = context.WithValue(ctx, dataKey, d)
	log.Info("reading!", zap.Any(string(dataKey), ctx.Value(dataKey)))

	// Simulate slow processing (replace with actual work)
	for i := 0; i < 6; i++ {
		select {
		case <-ctx.Done():
			log.Info("Context deadline reached during processing")
			return
		case <-time.After(1 * time.Second):
			// Simulate some work
			log.Info("Processing...", zap.Int("step", i+1))
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("reading!"))
}
