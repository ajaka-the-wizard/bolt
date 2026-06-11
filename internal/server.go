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
	"github.com/ajaka-the-wizard/bolt/internal/redis"
	"github.com/ajaka-the-wizard/bolt/internal/routes"
	"github.com/ajaka-the-wizard/bolt/internal/store"
	"github.com/ajaka-the-wizard/bolt/internal/workers"
	"github.com/gofiber/fiber/v3"
)

func Listen() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := slog.Default()
	env := configs.LoadEnv(logger)

	db := database.ConnectDB(ctx, logger, env.DATABASE_URL)
	defer db.CloseConn()

	rdb := redis.InitRedis(ctx, env, logger)
	store := store.InitStore(rdb, db)

	workers.InitInvoiceWorkers(ctx, store, logger)

	app := fiber.New()

	// Graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	Shutdown(sig, logger, app, db, rdb, cancel)

	api := app.Group("/api/v1", middlewares.GenerateUniqueId(), middlewares.LoggerMiddleware(), middlewares.LatencyCalculations(), middlewares.AuthMiddleware(env))

	routes.Route(api, store)

	err := app.Listen(env.PORT)
	if err != nil {
		logger.Error("Failed to bind to port", "port", env.PORT, "err", err)
		panic(err)
	}
}

// appShutdowner is the subset of *fiber.App used during graceful shutdown.
type appShutdowner interface {
	Shutdown() error
}

// redisConnCloser is the subset of *redis.Redis used during graceful shutdown.
type redisConnCloser interface {
	CloseConn() error
}

// dbConnCloser is the subset of *database.Repo used during graceful shutdown.
type dbConnCloser interface {
	CloseConn()
}

// doShutdown contains the testable core of Shutdown.
func doShutdown(sig <-chan os.Signal, logger *slog.Logger, app appShutdowner, db dbConnCloser, rdb redisConnCloser, cancel context.CancelFunc) {
	go func() {
		<-sig
		cancel()
		logger.Info("Shutting down server...")
		if err := app.Shutdown(); err != nil {
			logger.Error("Error during shutdown", "error", err.Error())
		}
		logger.Info("Closing redis connection")
		if err := rdb.CloseConn(); err != nil {
			logger.Error("Error closing redis connection", "error", err.Error())
		}
		logger.Info("Closing database connection")
		db.CloseConn()
	}()
}

func Shutdown(sig chan os.Signal, logger *slog.Logger, app *fiber.App, db *database.Repo, rdb *redis.Redis, cancel context.CancelFunc) {
	doShutdown(sig, logger, app, db, rdb, cancel)
}
