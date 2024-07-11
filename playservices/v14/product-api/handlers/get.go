package handlers

import (
	"net/http"

	"github.com/AmitSuresh/playground/playservices/v14/product-api/data"
	"go.uber.org/zap"
)

// swagger:route GET /products products listProducts
// Return a list of products from the database
// responses:
//	200: productsResponse

// ListAll handles GET requests and returns all current products
func (p *ProductsHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	//log, _ := zap.NewProduction()
	p.l.Debug("Handle GET Products")

	w.Header().Add("Content-Type", "application/json")

	curr := r.URL.Query().Get("currency")
	// fetch the products from the datastore
	lp, err := p.db.GetProducts(curr)
	if err != nil {
		p.l.Error("unable to fetch products", zap.Error(err))
	}
	// serialize the list to JSON
	err = data.ToJSON(lp, w)
	if err != nil {
		http.Error(w, "Unable to marshal json", http.StatusInternalServerError)
	}
}

// swagger:route GET /products/{id} products listSingleProduct
// Return a list of products from the database
// responses:
//	200: productResponse
//	404: errorResponse

// ListSingle handles GET requests
func (p *ProductsHandler) ListSingleProduct(rw http.ResponseWriter, r *http.Request) {
	id := getProductID(r)

	p.l.Debug("[DEBUG]", zap.Any("get record id ", id))

	curr := r.URL.Query().Get("currency")
	prod, err := p.db.GetProductByID(id, curr)

	switch err {
	case nil:

	case data.ErrProductNotFound:
		p.l.Error("fetching product ", zap.Error(err))

		rw.WriteHeader(http.StatusNotFound)
		data.ToJSON(&GenericError{Message: err.Error()}, rw)
		return
	default:
		p.l.Error("fetching product ", zap.Error(err))

		rw.WriteHeader(http.StatusInternalServerError)
		data.ToJSON(&GenericError{Message: err.Error()}, rw)
		return
	}

	err = data.ToJSON(prod, rw)
	if err != nil {
		// we should never be here but log the error just incase
		p.l.Error("serializing product", zap.Error(err))
	}
}
