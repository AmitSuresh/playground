package middleware

import (
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
)

func AddSwagger(app *fiber.App) {
	scfg := swagger.Config{
		BasePath: "/",
		FilePath: "app/docs/swagger.yaml",
		Path:     "docs",
		Title:    "Swagger API Docs",
	}

	app.Use(swagger.New(scfg))
}
