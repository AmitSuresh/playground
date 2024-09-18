package persistance

import (
	"errors"

	"github.com/AmitSuresh/playground/db-server/src/application/domain/entity"
	"gorm.io/gorm"
)

type orderRepository struct {
	db *gorm.DB
}

type OrderRepository interface {
	GetOrderById(id int64) (*entity.Order, error)
	CreateOrder(order entity.Order) (*entity.Order, error)
}

func (repo orderRepository) CreateOrder(order entity.Order) (*entity.Order, error) {
	tx := repo.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			print("panic occured during transaction: ", r)
		}
	}()

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &order, nil
}

func (repo orderRepository) GetOrderById(id int64) (*entity.Order, error) {
	tx := repo.db.Begin()
	var order entity.Order
	if err := tx.Preload("OrderLineItems").First(&order, id).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	return &order, nil
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}
