package middlewares

import (
	"log/slog"
	"time"

	"github.com/ajaka-the-wizard/bolt/internal/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// GenerateUniqueId generates an id for every requests, this aids observability and debugging
func GenerateUniqueId() fiber.Handler {
	return func(c fiber.Ctx) error {
		id := uuid.New().String()
		c.Locals("id", id)
		return c.Next()
	}
}

// This service uses a request scoped logger for easy request tracking. LoggerMiddleware is responsible for initializing it and injecting into the current request scope
func LoggerMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Locals("id").(string)
		logger := slog.Default().With(
			slog.String("request_id", id),
		)
		c.Locals("logger", logger)
		return c.Next()
	}
}

// LatencyCalculations is a small middleware to track the time taken for a request to complete. It aids observability
func LatencyCalculations() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		path := string(c.RequestCtx().Path())
		logger := utils.GetLogger(c)

		logger.Info("request started", "at", start, "method", string(c.RequestCtx().Method()), "path", path)
		c.Next()
		end := time.Now()
		latency := time.Since(start)
		status := c.RequestCtx().Response.StatusCode()

		msg := "request completed"
		attrs := []any{
			"status", status,
			"latency", latency,
			"path", path,
			"end", end,
		}
		if status >= 500 {
			logger.Error(msg, attrs...)
		} else if status >= 400 {
			logger.Warn(msg, attrs...)
		} else {
			logger.Info(msg, attrs...)
		}
		return nil
	}
}
