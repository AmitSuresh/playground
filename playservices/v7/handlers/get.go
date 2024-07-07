package handlers

import (
	"net/http"

	"github.com/AmitSuresh/playground/playservices/v7/data"
	"go.uber.org/zap"
)

// swagger:route GET /products products listProducts
// Return a list of products from the database
// responses:
//	200: productsResponse

// ListAll handles GET requests and returns all current products
func (p *ProductsHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	//log, _ := zap.NewProduction()
	p.l.Info("[INFO] Handle GET Products")

	// fetch the products from the datastore
	lp := data.GetProducts()

	// serialize the list to JSON
	err := data.ToJSON(lp, w)
	if err != nil {
		http.Error(w, "Unable to marshal json", http.StatusInternalServerError)
	}
}

// swagger:route GET /products/{id} products listSingle
// Return a list of products from the database
// responses:
//	200: productResponse
//	404: errorResponse

// ListSingle handles GET requests
func (p *ProductsHandler) ListSingle(rw http.ResponseWriter, r *http.Request) {
	id := getProductID(r)

	p.l.Info("[DEBUG]", zap.Any("get record id ", id))

	prod, err := data.GetProductByID(id)

	switch err {
	case nil:

	case data.ErrProductNotFound:
		p.l.Info("[ERROR]", zap.Any("fetching product ", err))

		rw.WriteHeader(http.StatusNotFound)
		data.ToJSON(&GenericError{Message: err.Error()}, rw)
		return
	default:
		p.l.Info("[ERROR]", zap.Any("fetching product ", err))

		rw.WriteHeader(http.StatusInternalServerError)
		data.ToJSON(&GenericError{Message: err.Error()}, rw)
		return
	}

	err = data.ToJSON(prod, rw)
	if err != nil {
		// we should never be here but log the error just incase
		p.l.Info("[ERROR]", zap.Any("serializing product ", err))
	}
}
