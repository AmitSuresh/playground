package handlers

import (
	"context"
	"net/http"
	"regexp"
	"strconv"

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

	switch r.Method {
	case http.MethodGet:
		p.getProducts(ctx)
	case http.MethodPost:
		p.addProducts(ctx)
	case http.MethodPut:
		p.updateProducts(ctx)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)

	}
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

func (p *Products) addProducts(ctx context.Context) {

	product := &data.Product{}

	l := GetLoggerFromContext(ctx)
	r := GetRequestFromContext(ctx)
	w := GetResponseWriterFromContext(ctx)

	l.Info("from addProducts")
	err := product.FromJSON(r.Body)
	if err != nil {
		http.Error(w, "Unable to unmarshal json", http.StatusInternalServerError)
	}
	l.Info("Parsed product", zap.Any("product", product))
	data.AddProduct(product)
}

func (p Products) updateProducts(ctx context.Context) {

	l := GetLoggerFromContext(ctx)
	r := GetRequestFromContext(ctx)
	w := GetResponseWriterFromContext(ctx)
	prod := &data.Product{}

	l.Info("from updateProducts")

	l.Info("PUT", zap.Any("", r.URL.Path))
	// expect the id in the URI
	reg := regexp.MustCompile(`/([0-9]+)`)
	g := reg.FindAllStringSubmatch(r.URL.Path, -1)

	if len(g) != 1 {
		l.Info("Invalid URI more than one id")
		http.Error(w, "Invalid URI", http.StatusBadRequest)
		return
	}

	if len(g[0]) != 2 {
		l.Info("Invalid URI more than one capture group")
		http.Error(w, "Invalid URI", http.StatusBadRequest)
		return
	}

	idString := g[0][1]
	id, err := strconv.Atoi(idString)
	if err != nil {
		l.Info("Invalid URI unable to convert to number", zap.Any("", idString))
		http.Error(w, "Invalid URI", http.StatusBadRequest)
		return
	}

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
