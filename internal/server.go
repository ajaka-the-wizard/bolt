package internal

import (
	"log/slog"

	"github.com/ajaka-the-wizard/bolt/internal/configs"
	v1 "github.com/ajaka-the-wizard/bolt/internal/routes/v1"
	"github.com/gofiber/fiber/v3"
)

func Listen() {
	logger := slog.Default()
	env := configs.LoadEnv(logger)

	app := fiber.New()
	api := app.Group("/api")

	v1.Route(api)

	app.Listen(env.PORT)
}
