package middlewares

import (
	"crypto/subtle"

	"github.com/ajaka-the-wizard/bolt/internal/configs"
	"github.com/ajaka-the-wizard/bolt/internal/store"
	"github.com/gofiber/fiber/v3"
)

func AuthMiddleware(env *configs.Env) fiber.Handler {
	return func(c fiber.Ctx) error {
		sharedSecret := c.Get("X-Shared-Secret")
		if sharedSecret == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "No secret provided"})
		}
		same := subtle.ConstantTimeCompare([]byte(sharedSecret), []byte(env.SHARED_SECRET))
		if same != 1 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Invalid secret provided"})
		}
		return c.Next()
	}
}

func IdempotencyMiddleware(s *store.Store) fiber.Handler {
	return func(c fiber.Ctx) error {
		idempotencyKey := c.Get("X-Idempotency-Key")
		if idempotencyKey == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "No idempotency key provided"})
		}
		if s.CheckKeyExistence(c.RequestCtx(), idempotencyKey) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Duplicate request detected"})
		}
		return c.Next()
	}
}
