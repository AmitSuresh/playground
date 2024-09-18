// Package classification db-server
//
// # Documentation for db-server
//
// Schemes: http
// BasePath: /
// Version: 1.0.0
//
// Consumes:
// - application/json
//
// Produces:
// - application/json
// swagger:meta
package doc

import (
	"github.com/AmitSuresh/playground/db-server/src/application/domain/entity"
	"github.com/AmitSuresh/playground/db-server/src/application/model"
)

//
// NOTE: Types defined here are purely for documentation purposes
// these types are not used by any of the handers

// Generic error message returned as a string
// swagger:response errorResponse
type errorResponseWrapper struct {
	// Description of the error
	// in: body
	Body model.GenericError
}

// Validation errors defined as an array of strings
// swagger:response validationErrorResponse
type errorValidationWrapper struct {
	// Collection of the errors
	// in: body
	Body model.ValidationError
}

// OrderResponse represents the response for an order
// swagger:response orderResponse
type OrderResponse struct {
	// An order in the system
	// in: body
	Body *entity.Order `json:"body"`
}

// swagger:parameters createOrder
type productParamsWrapper struct {
	// Product data structure to Update or Create.
	// in: body
	// required: true
	Body model.CreateOrderCommand
}
