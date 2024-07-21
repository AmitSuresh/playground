package handlers

import (
	"fmt"
	"net/http"

	"github.com/AmitSuresh/playground/playservices/v14/product-api/data"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	prod := GetProductsFromContext(ctx)

	i, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "invalid object id", http.StatusInternalServerError)
		p.l.Error("invalid object id", zap.Error(err))
	}

	p.l.Info("product id", zap.Any(string(productKey), i))
	p.l.Info("Handle PUT Products", zap.Any(string(logKey), ctx.Value(logKey)))
	p.l.Info("product from context", zap.Any(string(productKey), ctx.Value(productKey)))

	res, err := p.db.UpdateProduct(r.Context(), prod, i)

	if err != nil {
		switch err {
		case data.ErrProductNotFound:
			http.Error(w, fmt.Sprintf("product of id: %s not found in put ", i), http.StatusNotFound)
			p.l.Error("product not found in put", zap.Error(err))
			return
		default:
			http.Error(w, "product not found in put", http.StatusInternalServerError)
			p.l.Error("product not found in put", zap.Error(err))
			return
		}
	}

	err = data.ToJSON(res, w)
	if err != nil {
		p.l.Error("error serializing the contents", zap.Error(err))
	}

	// write the no content success header
	w.WriteHeader(http.StatusNoContent)
}
