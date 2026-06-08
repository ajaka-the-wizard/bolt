package handlers

import (
	"github.com/gofiber/fiber/v3"
)

func Letsgo() fiber.Handler {
	return func(c fiber.Ctx) error {
		return c.SendString("Lets go!!!!")
	}
}
