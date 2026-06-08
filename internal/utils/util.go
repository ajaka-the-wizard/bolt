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
