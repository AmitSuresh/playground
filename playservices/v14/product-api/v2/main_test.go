package main

/*
import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AmitSuresh/playground/playservices/v14/product-api/v2/sdk/client"
	"github.com/AmitSuresh/playground/playservices/v14/product-api/v2/sdk/client/products"
	"github.com/AmitSuresh/playground/playservices/v14/product-api/v2/sdk/models"
)

 test
go test -v
go test -coverprofile c.out
go tool cover -html c.out
go tool cover -html c.out -o coverage.html
go tool cover -func c.out

func TestPlayservicesClient(t *testing.T) {
	// Mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Mock response payload
		name := "asdf"
		description := "fsdfa"
		sku := "asdf-asa-aadfa"
		price := float32(55)

		payload := []*models.Product{
			{
				Name:        &name,
				Description: description,
				Price:       &price,
				SKU:         &sku,
			},
		}

		// Encode the payload to JSON and write it to the response
		json.NewEncoder(w).Encode(payload)
	}))
	defer mockServer.Close()

	// Configure the client to use the mock server
	cfg := client.DefaultTransportConfig().WithHost(mockServer.Listener.Addr().String())
	c := client.NewHTTPClientWithConfig(nil, cfg)

	params := products.NewListProductsParams()
	prod, err := c.Products.ListProducts(params)

	if err != nil {
		t.Fatal(err)
	}

	product := prod.GetPayload()[0]
	fmt.Printf("Name: %s, Description: %s, Price: %.2f, SKU: %s\n",
		*product.Name, product.Description, *product.Price, *product.SKU)
	//t.Fail()
}
*/
