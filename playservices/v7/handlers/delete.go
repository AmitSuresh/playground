package handlers

import (
	"net/http"

	"github.com/AmitSuresh/playground/playservices/v7/data"
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

	id := getProductID(r)

	p.l.Info("[DEBUG]", zap.Any("deleting record id", id))

	err := data.DeleteProduct(id)
	if err == data.ErrProductNotFound {
		p.l.Error("[ERROR]", zap.Any("deleting record id does not exist", err))

		w.WriteHeader(http.StatusNotFound)
		data.ToJSON(&GenericError{Message: err.Error()}, w)
		return
	}

	if err != nil {
		p.l.Error("[ERROR]", zap.Any("deleting record", err))

		w.WriteHeader(http.StatusInternalServerError)
		data.ToJSON(&GenericError{Message: err.Error()}, w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
