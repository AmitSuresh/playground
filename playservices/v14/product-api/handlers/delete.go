package handlers

import (
	"net/http"

	"github.com/AmitSuresh/playground/playservices/v14/product-api/data"
	"go.uber.org/zap"
)

// swagger:route DELETE /products/{id} products deleteProduct
// Update a products details
//
// responses:
//	201: noContentResponse
//  404: errorResponse
//  501: errorResponse

// Delete handles DELETE requests and removes items from the database
func (p *ProductsHandler) Delete(w http.ResponseWriter, r *http.Request) {

	id := p.getProductID(r)

	p.l.Info("deleting record", zap.Any("id:", id))

	err := p.db.DeleteProduct(r.Context(), id)
	if err == data.ErrProductNotFound {
		p.l.Error("deleting record id does not exist", zap.Error(err))

		w.WriteHeader(http.StatusNotFound)
		data.ToJSON(&GenericError{Message: err.Error()}, w)
		return
	}

	if err != nil {
		p.l.Error("deleting record", zap.Error(err))

		w.WriteHeader(http.StatusInternalServerError)
		data.ToJSON(&GenericError{Message: err.Error()}, w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
