package services

import (
	"github.com/AmitSuresh/playground/db-server/src/application/domain/entity"
	"github.com/AmitSuresh/playground/db-server/src/application/domain/persistance"
	"github.com/AmitSuresh/playground/db-server/src/application/model"
)

type orderService struct {
	orderRepository persistance.OrderRepository
}

type OrderService interface {
	CreateOrder(command model.CreateOrderCommand) (*entity.Order, error)
	GetOrderById(id int64) (*entity.Order, error)
}

func (service orderService) GetOrderById(id int64) (*entity.Order, error) {
	return service.orderRepository.GetOrderById(id)
}

func (service orderService) CreateOrder(command model.CreateOrderCommand) (*entity.Order, error) {
	order := model.MapToOrder(command)
	return service.orderRepository.CreateOrder(order)
}

func NewOrderService(orderRepository persistance.OrderRepository) OrderService {
	return &orderService{orderRepository: orderRepository}
}
