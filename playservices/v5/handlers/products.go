package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/AmitSuresh/playground/playservices/v5/data"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type ResponseWriterKeyType string

var rwKey ResponseWriterKeyType = "RESPONSE_WRITER"

type RequestKeyType string

var requestKey RequestKeyType = "REQUEST"

type ProductKeyType string

var productKey ProductKeyType = "REQUEST"

// Products is a http.Handler
type ProductsHandler struct {
	l *zap.Logger
}

// NewProducts creates a products handler with the given logger
func NewProducts(l *zap.Logger) *ProductsHandler {
	return &ProductsHandler{l}
}

// getProducts returns the products from the data store
func (p *ProductsHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
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

func (p *ProductsHandler) AddProduct(w http.ResponseWriter, r *http.Request) {
	p.l.Info("from addProducts")

	ctx := r.Context()
	ctx = InjectLogger(ctx, p.l)

	prod := GetProductFromContext(ctx)

	p.l.Info("Handle GET Products", zap.Any(string(logKey), ctx.Value(logKey)))
	p.l.Info("product from context", zap.Any(string(productKey), ctx.Value(productKey)))

	p.l.Info("from addProducts")
	data.AddProduct(prod)
}

func (p *ProductsHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "unable to convert id", http.StatusBadRequest)
	}
	ctx := r.Context()
	ctx = InjectLogger(ctx, p.l)

	prod := GetProductFromContext(ctx)
	p.l.Info("Handle PUT Products", zap.Any(string(logKey), ctx.Value(logKey)))
	p.l.Info("product from context", zap.Any(string(productKey), ctx.Value(productKey)))

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

func (p *ProductsHandler) MiddlewareValidateProduct(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			prod := &data.Product{}

			err := prod.FromJSON(r.Body)
			if err != nil {
				http.Error(w, "error reading product", http.StatusBadRequest)
				return
			}
			ctx := InjectProduct(r.Context(), prod)
			r = r.WithContext(ctx)

			p.l.Info("from middleware", zap.Any("request Info: ", loggableRequest(r)))

			next.ServeHTTP(w, r)
		})
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

func InjectProduct(ctx context.Context, prod *data.Product) context.Context {
	return context.WithValue(ctx, productKey, prod)
}

func GetProductFromContext(ctx context.Context) *data.Product {
	c, _ := ctx.Value(productKey).(*data.Product)
	return c
}
