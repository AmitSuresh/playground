package handlers

import (
	"net/http"
	"strconv"

	"github.com/AmitSuresh/playground/playservices/v14/product-api/data"
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

	id := r.URL.Query().Get("id")
	// fetch the product from the context
	prod := GetProductFromContext(ctx)

	i, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "error parsing id", http.StatusInternalServerError)
		p.l.Error("error parsing id", zap.Error(err))
	}
	prod.ID = i
	p.l.Info("product id", zap.Any(string(productKey), prod.ID))
	p.l.Info("Handle PUT Products", zap.Any(string(logKey), ctx.Value(logKey)))
	p.l.Info("product from context", zap.Any(string(productKey), ctx.Value(productKey)))

	err = p.db.UpdateProduct(prod)
	if err == data.ErrProductNotFound {
		http.Error(w, "product not found in put", http.StatusNotFound)
		p.l.Error("product not found in put", zap.Error(err))
		return
	}

	if err != nil {
		http.Error(w, "product not found in put", http.StatusInternalServerError)
		p.l.Error("product not found in put", zap.Error(err))
		return
	}

	// write the no content success header
	w.WriteHeader(http.StatusNoContent)
}
