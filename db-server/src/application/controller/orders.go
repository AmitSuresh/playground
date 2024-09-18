package controller

import (
	"fmt"
	"strconv"

	"github.com/AmitSuresh/playground/db-server/src/application/domain/services"
	"github.com/AmitSuresh/playground/db-server/src/application/model"
	"github.com/AmitSuresh/playground/db-server/src/infra/validation"
	"github.com/gofiber/fiber/v2"
)

// GetOrderById returns an order by id
// swagger:route GET /orders/{id} order listOrder
// Returns an Order
//
// Responses:
// 200: orderResponse
// 400: errorResponse
// 404: errorResponse
// 500: errorResponse
//
// Parameters:
//   + name: id
//     in: path
//     description: The ID of the order to retrieve
//     required: true
//     type: integer
//     format: int64
//   + name: x-correlationid
//     in: header
//     description: The correlation ID for tracking the request
//     required: true
//     type: string

// GetOrderById returns an order from the database based on id
func GetOrderById(app *fiber.App, orderService services.OrderService) fiber.Router {
	return app.Get("/orders/:id", func(ctx *fiber.Ctx) error {
		fmt.Printf("Your correlationId is %v", ctx.Locals("correlationId"))

		orderId := ctx.Params("id")
		id, err := strconv.ParseInt(orderId, 10, 64)

		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(&model.GenericError{Message: "Order id is not valid"})
		}

		order, err := orderService.GetOrderById(id)
		if err != nil {
			return ctx.Status(fiber.StatusNotFound).JSON(&model.GenericError{Message: "Order Not Found, sorry! :("})
		}

		if order == nil {
			return ctx.Status(fiber.StatusNotFound).JSON(&model.GenericError{Message: "Order Not Found, sorry! :("})
		}

		return ctx.Status(fiber.StatusOK).JSON(order)
	})
}

// CreateOrder handles the creation of an order.
// swagger:route POST /orders order createOrder
//
// Creates an Order
//
// Responses:
// 201: orderResponse
// 400: errorResponse
// 402: validationErrorResponse
// 500: errorResponse
//
// Parameters:
//   + name: x-correlationid
//     in: header
//     description: The correlation ID for tracking the request
//     required: true
//     type: string

// Creates an order
func CreateOrder(app *fiber.App, customValidator *validation.CustomValidator, orderService services.OrderService) fiber.Router {
	return app.Post("/orders", func(ctx *fiber.Ctx) error {
		var createOrderCommand model.CreateOrderCommand
		err := ctx.BodyParser(&createOrderCommand)
		if err != nil {
			return ctx.Status(fiber.StatusNotFound).JSON(&model.GenericError{Message: err.Error()})
		}

		err2, hasError := validateCreateOrderRequest(ctx, customValidator, createOrderCommand)
		if hasError {
			return err2
		}

		createdOrder, err := orderService.CreateOrder(createOrderCommand)
		if err != nil {
			return ctx.Status(fiber.StatusNotFound).JSON(&model.GenericError{Message: err.Error()})
		}

		return ctx.Status(fiber.StatusCreated).JSON(createdOrder)
	})
}

func validateCreateOrderRequest(ctx *fiber.Ctx, customValidator *validation.CustomValidator, request model.CreateOrderCommand) (error, bool) {
	if errs := customValidator.Validate(request); len(errs) > 0 && errs[0].HasError {
		return ctx.Status(fiber.StatusBadRequest).JSON(&model.ValidationError{Messages: errs.Errors()}), true
	}
	return nil, false
}
