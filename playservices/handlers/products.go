package handlers

import (
	"context"
	"net/http"

	"github.com/AmitSuresh/playground/playservices/data"
	"go.uber.org/zap"
)

type ResponseWriterKeyType string

var rwKey ResponseWriterKeyType = "RESPONSE_WRITER"

type RequestKeyType string

var requestKey ResponseWriterKeyType = "REQUEST"

// Products is a http.Handler
type Products struct {
	l *zap.Logger
}

// NewProducts creates a products handler with the given logger
func NewProducts(l *zap.Logger) *Products {
	return &Products{l}
}

// ServeHTTP is the main entry point for the handler and staisfies the http.Handler
// interface
func (p *Products) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	//log, _ := zap.NewProduction()
	log := p.l
	ctx = InjectLogger(ctx, log)
	log.Info("Handle GET Products", zap.Any(string(logKey), ctx.Value(logKey)))
	ctx = InjectResponseWriter(ctx, w)
	log.Info("rwKey", zap.Any(string(rwKey), ctx.Value(rwKey)))
	ctx = InjectRequest(ctx, r)
	log.Info("requestKey", zap.Any(string(requestKey), loggableRequest(r)))

	// handle the request for a list of products
	if r.Method == http.MethodGet {
		p.getProducts(ctx)
		return
	}

	// catch all
	// if no method is satisfied return an error
	w.WriteHeader(http.StatusMethodNotAllowed)
}

// getProducts returns the products from the data store
func (p *Products) getProducts(ctx context.Context) {

	// fetch the products from the datastore
	lp := data.GetProducts()

	l := GetLoggerFromContext(ctx)
	w := GetResponseWriterFromContext(ctx)
	l.Info("from getProducts")
	// serialize the list to JSON
	err := lp.ToJSON(w)
	if err != nil {
		http.Error(w, "Unable to marshal json", http.StatusInternalServerError)
	}
}

func loggableRequest(r *http.Request) map[string]interface{} {
	return map[string]interface{}{
		"Method":     r.Method,
		"URL":        r.URL.String(),
		"Header":     r.Header,
		"RemoteAddr": r.RemoteAddr,
		"UserAgent":  r.UserAgent(),
	}
}

func InjectRequest(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, requestKey, r)
}

func GetRequestFromContext(ctx context.Context) *http.Request {
	c, _ := ctx.Value(requestKey).(*http.Request)
	return c
}

func InjectResponseWriter(ctx context.Context, w http.ResponseWriter) context.Context {
	return context.WithValue(ctx, rwKey, w)
}

func GetResponseWriterFromContext(ctx context.Context) http.ResponseWriter {
	c, _ := ctx.Value(rwKey).(http.ResponseWriter)
	return c
}
