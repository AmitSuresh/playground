package handlers

import (
	"context"
	"net/http"
	"time"

	protos "github.com/AmitSuresh/playground/playservices/v14/currency"
	"github.com/AmitSuresh/playground/playservices/v14/product-api/v2/data"
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

	w.Header().Add("Content-Type", "application/json")

	// fetch the products from the datastore
	lp := data.GetProducts()

	// serialize the list to JSON
	err := data.ToJSON(lp, w)
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

	rr := &protos.RateRequest{
		Base:        protos.Currencies(protos.Currencies_value["EUR"]),
		Destination: protos.Currencies(protos.Currencies_value["GBP"]),
	}
	ctx, _ := context.WithTimeout(r.Context(), 10*time.Second)
	presp, err := p.cc.GetRate(ctx, rr)
	if err != nil {
		p.l.Error("[ERROR]", zap.Any("error getting new rate ", err))
		data.ToJSON(&GenericError{Message: err.Error()}, rw)
		return
	}

	prod.Price = presp.Rate * prod.Price

	err = data.ToJSON(prod, rw)
	if err != nil {
		// we should never be here but log the error just incase
		p.l.Info("[ERROR]", zap.Any("serializing product ", err))
	}
}
