package workers

import (
	"time"

	"github.com/ajaka-the-wizard/bolt/internal/models"
)

// ─── Layout constants ─────────────────────────────────────────────────────────

const (
	fontRegular = "regular"
	fontBold    = "bold"

	pageW = 595.28 // A4 in points
	pageH = 841.89

	mLeft  = 40.0
	mRight = 40.0
	contW  = pageW - mLeft - mRight // 515.28

	// Items table — column X positions
	colItem  = mLeft
	colQty   = 310.0
	colUnit  = 390.0
	colTotal = 480.0
	colEnd   = pageW - mRight // 555.28

	// Right-aligned block (invoice meta + totals)
	rightX = 370.0
)

// Simple pdf file name generator, could be improved
func generateOutputPath(o *models.Order) string {
	return "order" + o.OrderNumber + time.Now().String()
}
