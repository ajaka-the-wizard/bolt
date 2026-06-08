package internal

import (
	"log/slog"

	"github.com/ajaka-the-wizard/bolt/internal/configs"
	"github.com/ajaka-the-wizard/bolt/internal/middlewares"
	v1 "github.com/ajaka-the-wizard/bolt/internal/routes/v1"
	"github.com/gofiber/fiber/v3"
)

func Listen() {
	logger := slog.Default()
	env := configs.LoadEnv(logger)

	app := fiber.New()
	api := app.Group("/api", middlewares.GenerateUniqueId(), middlewares.LoggerMiddleware(), middlewares.LatencyCalculations())

	v1.Route(api)

	err := app.Listen(env.PORT)
	if err != nil {
		logger.Error("Failed to bind to port", "port", env.PORT, "err", err)
		panic(err)
	}
}
