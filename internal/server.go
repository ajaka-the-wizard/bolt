package internal

import (
	"context"
	"log/slog"

	"github.com/ajaka-the-wizard/bolt/internal/configs"
	"github.com/ajaka-the-wizard/bolt/internal/database"
	"github.com/ajaka-the-wizard/bolt/internal/middlewares"
	"github.com/ajaka-the-wizard/bolt/internal/routes"
	"github.com/gofiber/fiber/v3"
)

func Listen() {
	ctx := context.Background()
	logger := slog.Default()
	env := configs.LoadEnv(logger)

	db := database.ConnectDB(ctx, logger, env.DATABASE_URL)

	app := fiber.New()
	api := app.Group("/api/v1", middlewares.GenerateUniqueId(), middlewares.LoggerMiddleware(), middlewares.LatencyCalculations(), middlewares.AuthMiddleware(env))

	routes.Route(api, db)

	err := app.Listen(env.PORT)
	if err != nil {
		logger.Error("Failed to bind to port", "port", env.PORT, "err", err)
		panic(err)
	}
}
