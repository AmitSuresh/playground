package model

import "github.com/AmitSuresh/playground/db-server/src/application/domain/entity"

type CreateOrderCommand struct {
	// shipment no of Order
	ShipmentNumber int64 `json:"shipmentNumber" validate:"required"`
	// cargo id of Order
	CargoId int `json:"cargoId" validate:"required"`
	// cargo id of Order
	OrderLineItems []CreateOrderLineItemCommand `json:"lineItems" validate:"required"`
}

type CreateOrderLineItemCommand struct {
	// product id of Order line items
	ProductId int64 `json:"productId" validate:"required"`
	// product id of Order line items
	SellerId int64 `json:"sellerId" validate:"required"`
}

// ValidationError is a collection of validation error messages
type ValidationError struct {
	Messages []string `json:"messages"`
}

// GenericError is a generic error message returned by a server
type GenericError struct {
	Message string `json:"message"`
}

func MapToOrder(command CreateOrderCommand) entity.Order {
	var items []entity.OrderLineItem

	if command.OrderLineItems != nil {
		for _, lineItemCommand := range command.OrderLineItems {
			items = append(items, entity.OrderLineItem{
				SellerId:  lineItemCommand.SellerId,
				ProductId: lineItemCommand.ProductId,
			})
		}
	}

	return entity.Order{
		IsShipped:      false,
		CargoId:        command.CargoId,
		ShipmentNumber: command.ShipmentNumber,
		OrderLineItems: items,
	}
}
