package utils

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
)

func GetLogger(c fiber.Ctx) *slog.Logger {
	if logger, ok := c.Locals("logger").(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

func GetKey(c fiber.Ctx, key string) (string, bool) {
	val, ok := c.Locals(key).(string)
	if !ok || val == "" {
		return "", false
	}
	return val, true
}
