package routes

import (
	"github.com/ajaka-the-wizard/bolt/internal/database"
	"github.com/ajaka-the-wizard/bolt/internal/handlers"
	"github.com/ajaka-the-wizard/bolt/internal/middlewares"
	"github.com/gofiber/fiber/v3"
)

func Route(api fiber.Router, r *database.Repo) {
	queue := api.Group("/queue")
	queue.Post("/", middlewares.IdempotencyMiddleware(), handlers.ProducerHandler(r))
}
