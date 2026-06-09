package handlers

import (
	"net/http"

	"github.com/ajaka-the-wizard/bolt/internal/models"
	"github.com/ajaka-the-wizard/bolt/internal/store"
	"github.com/ajaka-the-wizard/bolt/internal/utils"
	"github.com/gofiber/fiber/v3"
)

func ProducerHandler(s *store.Store) fiber.Handler {
	return func(c fiber.Ctx) error {
		// logger := utils.GetLogger(c)
		key, ok := utils.GetKey(c, "iKey")
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Something went wrong"})
		}
		var order models.Order
		if err := c.Bind().Body(&order); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
		}
		id, err := s.SaveOrder(c.RequestCtx(), &order, key)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to persist order"})
		}
		return c.Status(http.StatusAccepted).JSON(fiber.Map{"success": true, "message": "Job received for processing", "id": id})
	}
}
