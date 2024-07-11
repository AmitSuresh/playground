package data

import (
	"context"
	"fmt"

	protos "github.com/AmitSuresh/playground/playservices/v14/currency/protos/currency"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrProductNotFound = fmt.Errorf("product not found")

// Product defines the structure for an API product
type Product struct {
	// the id for the product
	//
	// required: false
	// min: 1
	ID int `json:"id"` // Unique identifier for the product
	// the name for this poduct
	//
	// required: true
	// max length: 255
	Name string `json:"name" validate:"required"`
	// the description for this poduct
	//
	// required: false
	// max length: 10000
	Description string `json:"description"`
	// the price for the product
	//
	// required: true
	// min: 0.01
	Price float64 `json:"price" validate:"gt=0"`
	// the SKU for the product
	//
	// required: true
	// pattern: [a-z]+-[a-z]+-[a-z]+
	SKU string `json:"sku" validate:"required,sku"`
}

// Products is a collection of Product
type Products []*Product

type ProductsDB struct {
	currencyClient protos.CurrencyClient
	l              *zap.Logger
	rates          map[string]float64
	currSubClient  protos.Currency_SubscribeRatesClient
}

func GetProductsDB(c protos.CurrencyClient, l *zap.Logger) *ProductsDB {
	db := &ProductsDB{c, l, make(map[string]float64), nil}

	go db.handleUpdates()

	return db
}

func (db *ProductsDB) handleUpdates() {
	// Recv returns a StreamingRateResponse which can contain one of two messages
	// RateResponse or an Error.
	// We need to handle each case separately
	subClient, err := db.currencyClient.SubscribeRates(context.Background())

	if err != nil {
		// handle connection errors
		// this is normally terminal requires a reconnect
		db.l.Error("unable to subscribe for rates", zap.Error(err))
		return
	}

	db.currSubClient = subClient

	for {
		// Recv returns a StreamingRateResponse which can contain one of two messages
		// RateResponse or an Error.
		// We need to handle each case separately
		sresp, err := db.currSubClient.Recv()

		// handle connection errors
		// this is normally terminal requires a reconnect
		if err != nil {
			db.l.Error("error receiving message", zap.Error(err))
			return
		}

		// handle a returned error message
		if ss := sresp.GetError(); ss != nil {
			db.l.Error("error subscribing for rates", zap.Any("error", ss))
			sre := status.FromProto(ss)
			if sre.Code() == codes.InvalidArgument {
				errDetails := ""
				// get the RateRequest serialized in the error response
				// Details is a collection but we are only returning a single item
				if d := sre.Details(); len(d) > 0 {
					db.l.Error("", zap.Any("details", d))
					if rr, ok := d[0].(*protos.RateRequest); ok {
						errDetails = fmt.Sprintf("base: %s destination: %s", rr.GetBase().String(), rr.GetDestination().String())
					}
				}
				db.l.Error("error receiving message", zap.Any("", errDetails))
			}
		}

		// handle a rate response
		if rresp := sresp.GetRateResponse(); rresp != nil {
			db.l.Info("received updated rate from server", zap.Any("destination", rresp.Destination.String()))
			db.rates[rresp.Destination.String()] = rresp.Rate
		}
	}
}

// GetProducts returns all products from the database
func (db *ProductsDB) GetProducts(currency string) (Products, error) {

	if currency == "" {
		return productList, nil
	}
	r, err := db.getRate(currency)
	if err != nil {
		db.l.Error("[ERROR] unable to get rate", zap.Any("currency", currency), zap.Error(err))
		return nil, err
	}

	products := Products{}
	for _, p := range productList {
		np := *p
		np.Price = np.Price * r
		products = append(products, &np)
	}

	return products, nil
}

// GetProductByID returns a single product which matches the id from the
// database.
// If a product is not found this function returns a ProductNotFound error
func (db *ProductsDB) GetProductByID(id int, currency string) (*Product, error) {
	i := findIndexByProductID(id)
	if id == -1 {
		return nil, ErrProductNotFound
	}

	if currency == "" {
		return productList[i], nil
	}

	r, err := db.getRate(currency)
	if err != nil {
		db.l.Error("[ERROR] unable to get rate", zap.Any("currency", currency), zap.Error(err))
		return nil, err
	}

	newp := *productList[i]
	newp.Price = newp.Price * r
	return &newp, nil
}

// AddProduct adds a new product to the database
func (db *ProductsDB) AddProduct(p *Product) {
	// get the next id in sequence
	maxID := productList[len(productList)-1].ID
	p.ID = maxID + 1
	productList = append(productList, p)
}

// UpdateProduct replaces a product in the database with the given
// item.
// If a product with the given id does not exist in the database
// this function returns a ProductNotFound error
func (db *ProductsDB) UpdateProduct(p *Product) error {
	i := findIndexByProductID(p.ID)
	if i == -1 {
		return ErrProductNotFound
	}

	// update the product in the DB
	productList[i] = p

	return nil
}

// DeleteProduct deletes a product from the database
func (db *ProductsDB) DeleteProduct(id int) error {
	i := findIndexByProductID(id)
	if i == -1 {
		return ErrProductNotFound
	}

	productList = append(productList[:i], productList[i+1])

	return nil
}

// findIndex finds the index of a product in the database
// returns -1 when no product can be found
func findIndexByProductID(id int) int {
	for i, p := range productList {
		if p.ID == id {
			return i
		}
	}

	return -1
}

func (db *ProductsDB) getRate(destination string) (float64, error) {
	// if cached return
	/* 	if r, ok := db.rates[destination]; ok {
		return r, nil
	} */

	req := &protos.RateRequest{
		Base:        protos.Currencies(protos.Currencies_value["EUR"]),
		Destination: protos.Currencies(protos.Currencies_value[destination]),
	}

	// get initial rate
	resp, err := db.currencyClient.GetRate(context.Background(), req)
	if err != nil {
		if s, ok := status.FromError(err); ok {
			md := s.Details()[0].(*protos.RateRequest)
			if s.Code() == codes.InvalidArgument {
				return -1, fmt.Errorf("base %v and destination currencies %v cannot be the same", md.Base.String(), md.Destination.String())
			}
			return -1, fmt.Errorf("unable to get rate from currency server for Base: %v, Destination: %v", md.Base.String(), md.Destination.String())
		}
	}
	db.rates[destination] = resp.Rate

	// subscribe for updates
	db.currSubClient.Send(req)
	if err != nil {
		return -1, err
	}

	return resp.Rate, err
}

// productList is a hard coded list of products for this
// example data source
var productList = []*Product{
	{
		ID:          1,
		Name:        "Latte",
		Description: "Frothy milky coffee",
		Price:       2.45,
		SKU:         "abc323",
	},
	{
		ID:          2,
		Name:        "Espresso",
		Description: "Short and strong coffee without milk",
		Price:       1.99,
		SKU:         "fjd34",
	},
}
