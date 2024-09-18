package main

import (
	"fmt"

	"github.com/AmitSuresh/playground/db-server/src/application/controller"
	"github.com/AmitSuresh/playground/db-server/src/application/domain/entity"
	"github.com/AmitSuresh/playground/db-server/src/application/domain/persistance"
	"github.com/AmitSuresh/playground/db-server/src/application/domain/services"
	"github.com/AmitSuresh/playground/db-server/src/infra/config"
	"github.com/AmitSuresh/playground/db-server/src/infra/middleware"
	"github.com/AmitSuresh/playground/db-server/src/infra/validation"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	app *fiber.App
	l   *zap.Logger
	cfg *config.Config
	db  *gorm.DB
	//d   bool
)

func init() {

	l, _ = zap.NewProduction()
	app = fiber.New()
	config, err := config.LoadConfig(l)
	if err != nil {
		l.Error("\nError loding config")
	}
	cfg = config

	database, err := gorm.Open(postgres.Open(cfg.DBUrl), &gorm.Config{})
	if err != nil {
		l.Error("\nfailed to connect to database.", zap.Any("cfg.DBUrl", cfg.DBUrl))
		//d = false
	}
	db = database

	db.AutoMigrate(&entity.Order{}, &entity.OrderLineItem{})
	customValidator := validation.NewValidation()

	app.Use(recover.New())
	app.Use(cors.New())

	// middleware
	middleware.AddCorrelationId(app)
	middleware.AddSwagger(app)

	// repositories
	orderRepo := persistance.NewOrderRepository(db)

	// services
	orderService := services.NewOrderService(orderRepo)

	// endpoints
	controller.GetOrderById(app, orderService)
	controller.CreateOrder(app, customValidator, orderService)
}

func main() {
	defer l.Sync()
	l.Info("intialized successfully!", zap.Any("\nWill start listening on: ", fmt.Sprintf("%s:%s", cfg.ServerAddr, cfg.ServerPort)))
	app.Listen(fmt.Sprintf("%s:%s", cfg.ServerAddr, cfg.ServerPort))
}
