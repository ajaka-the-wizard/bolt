package handlers

import (
	"net/http"

	"github.com/ajaka-the-wizard/bolt/internal/database"
	"github.com/ajaka-the-wizard/bolt/internal/models"
	"github.com/gofiber/fiber/v3"
)

func ProducerHandler(r *database.Repo) fiber.Handler {
	return func(c fiber.Ctx) error {
		// logger := utils.GetLogger(c)
		var order models.Order
		if err := c.Bind().Body(&order); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
		}

		id := r.SaveOrder(c.RequestCtx(), &order)
		return c.Status(http.StatusAccepted).JSON(fiber.Map{"Success": true, "message": "Job received for processing", "id": id})
	}
}
