package handlers

import (
	"fmt"
	"net/http"

	"github.com/AmitSuresh/playground/playservices/v14/product-api/data"
	"go.uber.org/zap"
)

// swagger:route POST /products products createProduct
// Create a new product
//
// responses:
//	200: productResponse
//  422: errorValidation
//  501: errorResponse

// Create handles POST requests to add new products
func (p *ProductsHandler) Create(w http.ResponseWriter, r *http.Request) {

	// fetch the product from the context
	prod := GetProductsFromContext(r.Context())

	p.l.Info("inserting a new product", zap.Any("", prod))

	var docs []interface{}
	for _, p := range prod {
		docs = append(docs, p)
	}
	res, err := p.db.AddProduct(r.Context(), docs)
	if err != nil {
		p.l.Error("error creating a new product", zap.Error(err))
		data.ToJSON(&GenericError{Message: err.Error()}, w)
	}
	data.ToJSON(fmt.Sprintf("inserted id: %s", res), w)

	w.WriteHeader(http.StatusOK)
}
