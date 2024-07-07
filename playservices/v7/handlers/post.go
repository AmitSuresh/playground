package handlers

import (
	"net/http"

	"github.com/AmitSuresh/playground/playservices/v7/data"
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
	prod := GetProductFromContext(r.Context())

	p.l.Info("[DEBUG]", zap.Any("inserting a new product", prod))

	data.AddProduct(prod)

	data.ToJSON("inserted a new product successfully", w)
	w.WriteHeader(http.StatusOK)
}
