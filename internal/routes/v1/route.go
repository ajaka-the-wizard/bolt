package v1

import (
	"github.com/ajaka-the-wizard/bolt/internal/handlers"
	"github.com/gofiber/fiber/v3"
)

func Route(api fiber.Router) {
	v1 := api.Group("/v1")
	v1.Get("/", handlers.Letsgo())
}
