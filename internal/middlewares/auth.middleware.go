package middlewares

import (
	"crypto/subtle"

	"github.com/ajaka-the-wizard/bolt/internal/configs"
	"github.com/ajaka-the-wizard/bolt/internal/store"
	"github.com/ajaka-the-wizard/bolt/internal/utils"
	"github.com/gofiber/fiber/v3"
)

func AuthMiddleware(env *configs.Env) fiber.Handler {
	return func(c fiber.Ctx) error {
		logger := utils.GetLogger(c)
		sharedSecret := c.Get("X-Shared-Secret")
		if sharedSecret == "" {
			logger.Warn("Request does not include required sharedsecret")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "No secret provided"})
		}
		same := subtle.ConstantTimeCompare([]byte(sharedSecret), []byte(env.SHARED_SECRET))
		if same != 1 {
			logger.Warn("Request provided invalid secret")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Invalid secret provided"})
		}
		return c.Next()
	}
}

func IdempotencyMiddleware(s *store.Store) fiber.Handler {
	return func(c fiber.Ctx) error {
		logger := utils.GetLogger(c)
		idempotencyKey := c.Get("X-Idempotency-Key")
		if idempotencyKey == "" {
			logger.Warn("Idempotency key not provided")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "No idempotency key provided"})
		}
		c.Locals("iKey", idempotencyKey)
		if s.CheckKeyExistence(c.RequestCtx(), idempotencyKey) {
			logger.Warn("Duplicate request detected")
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Duplicate request detected"})
		}
		if err := s.SetIdempotencyKey(c.RequestCtx(), idempotencyKey); err != nil {
			logger.Error("Error setting idempotency key to database", "error", err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Something went wrong"})
		}
		return c.Next()
	}
}
