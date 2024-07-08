package handlers

import (
	"net/http"

	"github.com/AmitSuresh/playground/playservices/v7/data"
	"go.uber.org/zap"
)

// swagger:route PUT /products products updateProduct
// Update a products details
//
// responses:
//	201: noContentResponse
//  404: errorResponse
//  422: errorValidation

// Update handles PUT requests to update products
func (p *ProductsHandler) Update(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	ctx = InjectLogger(ctx, p.l)

	// fetch the product from the context
	prod := GetProductFromContext(ctx)
	p.l.Info("[DEBUG] Handle PUT Products", zap.Any(string(logKey), ctx.Value(logKey)))
	p.l.Info("[DEBUG] product from context", zap.Any(string(productKey), ctx.Value(productKey)))

	err := data.UpdateProduct(prod)
	if err == data.ErrProductNotFound {
		http.Error(w, "product not found", http.StatusNotFound)
		p.l.Error("[ERROR]", zap.Any("product not found", err))
		return
	}

	if err != nil {
		http.Error(w, "product not found", http.StatusInternalServerError)
		return
	}

	// write the no content success header
	w.WriteHeader(http.StatusNoContent)
}
