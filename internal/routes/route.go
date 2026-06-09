package routes

import (
	"github.com/ajaka-the-wizard/bolt/internal/handlers"
	"github.com/ajaka-the-wizard/bolt/internal/middlewares"
	"github.com/ajaka-the-wizard/bolt/internal/store"
	"github.com/gofiber/fiber/v3"
)

func Route(api fiber.Router, s *store.Store) {
	queue := api.Group("/queue")
	queue.Post("/", middlewares.IdempotencyMiddleware(s), handlers.ProducerHandler(s))
}
