package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/AmitSuresh/playground/playservices/data"
	"github.com/gorilla/mux"
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

// getProducts returns the products from the data store
func (p *Products) GetProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	//log, _ := zap.NewProduction()
	ctx = InjectLogger(ctx, p.l)
	p.l.Info("Handle GET Products", zap.Any(string(logKey), ctx.Value(logKey)))

	// fetch the products from the datastore
	lp := data.GetProducts()

	p.l.Info("from getProducts")
	// serialize the list to JSON
	err := lp.ToJSON(w)
	if err != nil {
		http.Error(w, "Unable to marshal json", http.StatusInternalServerError)
	}
}

func (p *Products) AddProduct(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	product := &data.Product{}

	//log, _ := zap.NewProduction()
	ctx = InjectLogger(ctx, p.l)
	p.l.Info("Handle GET Products", zap.Any(string(logKey), ctx.Value(logKey)))

	p.l.Info("from addProducts")
	err := product.FromJSON(r.Body)
	if err != nil {
		http.Error(w, "Unable to unmarshal json", http.StatusInternalServerError)
	}
	p.l.Info("Parsed product", zap.Any("product", product))
	data.AddProduct(product)
}

func (p Products) UpdateProduct(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "unable to convert id", http.StatusBadRequest)
	}
	ctx := r.Context()
	//log, _ := zap.NewProduction()
	ctx = InjectLogger(ctx, p.l)
	p.l.Info("Handle PUT Products", zap.Any(string(logKey), ctx.Value(logKey)))

	p.l.Info("from updateProducts")
	prod := &data.Product{}

	err = prod.FromJSON(r.Body)
	if err != nil {
		http.Error(w, "Unable to unmarshal json", http.StatusBadRequest)
	}

	err = data.UpdateProduct(id, prod)
	if err == data.ErrProductNotFound {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Product not found", http.StatusInternalServerError)
		return
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
