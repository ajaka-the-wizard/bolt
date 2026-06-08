package middlewares

import (
	"log/slog"
	"time"

	"github.com/ajaka-the-wizard/bolt/internal/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func GenerateUniqueId() fiber.Handler {
	return func(c fiber.Ctx) error {
		id := uuid.New().String()
		c.Locals("id", id)
		return c.Next()
	}
}

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
