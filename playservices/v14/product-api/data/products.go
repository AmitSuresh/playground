package data

import (
	"context"
	"fmt"
	"log"

	protos "github.com/AmitSuresh/playground/playservices/v14/currency/protos/currency"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var ErrProductNotFound = fmt.Errorf("product not found")

// Product defines the structure for an API product
type Product struct {
	// the id for the product
	//
	// required: false
	// min: 1
	ID string `json:"id,omitempty" bson:"_id,omitempty"`
	// the name for this poduct
	//
	// required: true
	// max length: 255
	Name string `json:"name" bson:"name" validate:"required"`
	// the description for this poduct
	//
	// required: false
	// max length: 10000
	Description string `json:"description" bson:"description,omitempty"`
	// the price for the product
	//
	// required: true
	// min: 0.01
	Price float64 `json:"price" bson:"price" validate:"gt=0"`
	// the SKU for the product
	//
	// required: true
	// pattern: [a-z]+-[a-z]+-[a-z]+
	SKU string `json:"sku" bson:"sku" validate:"required,sku"`
}

// Products is a collection of Product
type Products []*Product

type ProductsDB struct {
	currencyClient  protos.CurrencyClient
	l               *zap.Logger
	rates           map[string]float64
	currSubClient   protos.Currency_SubscribeRatesClient
	mongoClient     *mongo.Client
	mongoCollection *mongo.Collection
}

func GetProductsDB(c protos.CurrencyClient, l *zap.Logger, mc *mongo.Client) *ProductsDB {
	db := &ProductsDB{c, l, make(map[string]float64), nil, nil, nil}

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
func (db *ProductsDB) GetProducts(ctx context.Context, currency string) (Products, error) {

	db.l.Info("here")
	if err := db.mongoClient.Ping(ctx, nil); err != nil {
		db.l.Error("mongoClient is not connected", zap.Error(err))
		return nil, fmt.Errorf("client is disconnected: %v", err)
	}

	filter := bson.M{}
	cursor, err := db.mongoCollection.Find(ctx, filter)
	if err != nil {
		db.l.Error("error finding data", zap.Error(err))
	}
	var results []*Product

	if err = cursor.All(ctx, &results); err != nil {
		db.l.Error("error decoding data", zap.Error(err))
	}
	if currency == "" {
		return results, nil
	}
	db.l.Info("here", zap.Any("here", results))
	r, err := db.getRate(currency)
	if err != nil {
		db.l.Error("[ERROR] unable to get rate", zap.Any("currency", currency), zap.Error(err))
		return nil, err
	}

	for _, v := range results {
		v.Price = v.Price * r
	}

	/* 	products := Products{}
	   	for _, p := range productList {
	   		np := *p
	   		np.Price = np.Price * r
	   		products = append(products, &np)
	   	}
	*/
	return results, nil
}

// GetProductByID returns a single product which matches the id from the
// database.
// If a product is not found this function returns a ProductNotFound error
func (db *ProductsDB) GetProductByID(ctx context.Context, id primitive.ObjectID, currency string) (*Product, error) {

	filter := bson.D{
		{
			Key:   "_id",
			Value: id,
		},
	}
	p := new(Product)
	err := db.mongoCollection.FindOne(ctx, filter).Decode(p)
	if err != nil {
		db.l.Error("[ERROR] unable to find the product", zap.Error(err))
		return nil, err
	}

	if currency == "" {
		return p, nil
	}

	r, err := db.getRate(currency)
	if err != nil {
		db.l.Error("[ERROR] unable to get rate", zap.Any("currency", currency), zap.Error(err))
		return nil, err
	}

	p.Price = p.Price * r
	return p, nil
}

// AddProduct adds a new product to the database
func (db *ProductsDB) AddProduct(ctx context.Context, d []interface{}) (*mongo.InsertManyResult, error) {

	res, err := db.mongoCollection.InsertMany(ctx, d)
	if err != nil {
		return nil, fmt.Errorf("error inserting records: %v", err)
	}
	return res, nil
}

// UpdateProduct replaces a product in the database with the given
// item.
// If a product with the given id does not exist in the database
// this function returns a ProductNotFound error
func (db *ProductsDB) UpdateProduct(ctx context.Context, p []*Product, id primitive.ObjectID) (*mongo.BulkWriteResult, error) {

	var updateModels []mongo.WriteModel
	filter := bson.D{
		{
			Key:   "_id",
			Value: id,
		},
	}

	for _, prod := range p {
		update := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "name", Value: prod.Name},
				{Key: "description", Value: prod.Description},
				{Key: "price", Value: prod.Price},
				{Key: "sku", Value: prod.SKU},
			}},
		}
		model := mongo.NewUpdateManyModel().SetFilter(filter).SetUpdate(update)
		updateModels = append(updateModels, model)
	}

	res, err := db.mongoCollection.BulkWrite(ctx, updateModels)
	if err != nil {
		db.l.Error("error updating one product", zap.Error(err))
		return nil, err
	}

	switch res.MatchedCount {
	case 0:
		return nil, ErrProductNotFound
	default:
		return res, nil
	}
}

// DeleteProduct deletes a product from the database
func (db *ProductsDB) DeleteProduct(ctx context.Context, id primitive.ObjectID) error {

	filter := bson.D{
		{
			Key:   "_id",
			Value: id,
		},
	}
	res, err := db.mongoCollection.DeleteOne(ctx, filter)
	if err != nil {
		db.l.Error("error deleting product from database", zap.Error(err))
		return err
	}

	if res.DeletedCount == 0 {
		return ErrProductNotFound
	}

	return nil
}

func (db *ProductsDB) DisconnectMongoClient() error {

	return db.mongoClient.Disconnect(context.Background())
}

func (db *ProductsDB) GetMongoClient(m string) (*mongo.Client, error) {

	s := options.ServerAPI(options.ServerAPIVersion1)
	ops := options.Client().ApplyURI(m).SetServerAPIOptions(s)

	// Create a new client and connect to the server
	mongoClient, err := mongo.Connect(context.Background(), ops)
	if err != nil {
		db.mongoClient = nil
		db.l.Error("error creating a new client")
		return nil, err
	}

	db.mongoClient = mongoClient
	return mongoClient, nil
}

func (db *ProductsDB) GetMongoCollection(dbase string, coll string) error {
	db.mongoCollection = db.mongoClient.Database(dbase).Collection(coll)
	if db.mongoCollection == nil {
		return fmt.Errorf("error retrieving collection")
	}
	return nil
}

func (db *ProductsDB) MigrateDocs(ctx context.Context) (*mongo.InsertManyResult, error) {
	newProd := []interface{}{
		Product{
			Name:        "Latte",
			Description: "Frothy milky coffee",
			Price:       2.45,
			SKU:         "abc323",
		},
		Product{
			Name:        "Espresso",
			Description: "Short and strong coffee without milk",
			Price:       1.99,
			SKU:         "fjd34",
		},
	}

	result, err := db.mongoCollection.InsertMany(ctx, newProd)
	if err != nil {
		db.l.Error("error migrating", zap.Error(err))
		return nil, err
	}
	return result, nil
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

func GetgrpcClient(s string, l *zap.Logger) *grpc.ClientConn {

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient(s, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	//defer conn.Close()

	return conn
}

// productList is a hard coded list of products for this
// example data source
var productList = []*Product{
	{
		//ID:          "1",
		//ID:          "669d04756b448f4c1fef495e"
		Name:        "Latte",
		Description: "Frothy milky coffee",
		Price:       2.45,
		SKU:         "abc323",
	},
	{
		//ID:          "2",
		//ID:          "669d04756b448f4c1fef495f",
		Name:        "Espresso",
		Description: "Short and strong coffee without milk",
		Price:       1.99,
		SKU:         "fjd34",
	},
}

var ProductList = productList
