// Package classification of Playservices API
//
// # Documentation of Playservices API
//
// Schemes: http
// BasePath: /products
// Version: 6.0.0
//
// Consumes:
// - application/json
//
// Produces:
// -application/json
// swagger:meta
package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/AmitSuresh/playground/playservices/v7/data"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type ResponseWriterKeyType string

var rwKey ResponseWriterKeyType = "RESPONSE_WRITER"

type RequestKeyType string

var requestKey RequestKeyType = "REQUEST"

type ProductKeyType string

var productKey ProductKeyType = "REQUEST"

type LoggerKeyType string

var logKey LoggerKeyType = "LOGGER"

// ErrInvalidProductPath is an error message when the product path is not valid
var ErrInvalidProductPath = fmt.Errorf("invalid Path, path should be /products/[id]")

// GenericError is a generic error message returned by a server
type GenericError struct {
	Message string `json:"message"`
}

// ValidationError is a collection of validation error messages
type ValidationError struct {
	Messages []string `json:"messages"`
}

// Products is a http.Handler
type ProductsHandler struct {
	l *zap.Logger
	v *data.Validation
}

// NewProducts returns a new products handler with the given logger
func NewProducts(l *zap.Logger, v *data.Validation) *ProductsHandler {
	return &ProductsHandler{l, v}
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

// getProductID returns the product ID from the URL
// Panics if cannot convert the id into an integer
// this should never happen as the router ensures that
// this is a valid number
func getProductID(r *http.Request) int {
	// parse the product id from the url
	vars := mux.Vars(r)

	// convert the id into an integer and return
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		// should never happen
		panic(err)
	}

	return id
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

func InjectLogger(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, logKey, l)
}

func GetLoggerFromContext(ctx context.Context) *zap.Logger {
	c, _ := ctx.Value(logKey).(*zap.Logger)
	return c
}
