package internal

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ajaka-the-wizard/bolt/internal/configs"
	"github.com/ajaka-the-wizard/bolt/internal/database"
	"github.com/ajaka-the-wizard/bolt/internal/middlewares"
	"github.com/ajaka-the-wizard/bolt/internal/queues"
	"github.com/ajaka-the-wizard/bolt/internal/redis"
	"github.com/ajaka-the-wizard/bolt/internal/routes"
	"github.com/ajaka-the-wizard/bolt/internal/store"
	"github.com/gofiber/fiber/v3"
)

func Listen() {
	ctx := context.Background()
	logger := slog.Default()
	env := configs.LoadEnv(logger)

	db := database.ConnectDB(ctx, logger, env.DATABASE_URL)
	defer db.CloseConn()

	rdb := redis.InitRedis(ctx, env, logger)
	defer func() {
		if err := rdb.CloseConn(); err != nil {
			logger.Error("Error closing redis connection", "error", err.Error())
		}
	}()
	queue := queues.InitQueue()
	store := store.InitStore(rdb, db, queue)
	app := fiber.New()

	// Graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	Shutdown(sig, logger, app, db, rdb)

	api := app.Group("/api/v1", middlewares.GenerateUniqueId(), middlewares.LoggerMiddleware(), middlewares.LatencyCalculations(), middlewares.AuthMiddleware(env))

	routes.Route(api, store)

	err := app.Listen(env.PORT)
	if err != nil {
		logger.Error("Failed to bind to port", "port", env.PORT, "err", err)
		panic(err)
	}
}

func Shutdown(sig chan os.Signal, logger *slog.Logger, app *fiber.App, db *database.Repo, rdb *redis.Redis) {
	go func() {
		<-sig
		logger.Info("Closing redis connection")
		if err := rdb.CloseConn(); err != nil {
			logger.Error("Error closing redis connection", "error", err.Error())
		}
		logger.Info("Closing database connection")
		db.CloseConn()
		logger.Info("Shutting down server...")
		if err := app.Shutdown(); err != nil {
			logger.Error("Error during shutdown", "error", err.Error())
		}
	}()
}
