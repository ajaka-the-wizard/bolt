package workers

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/ajaka-the-wizard/bolt/internal/models"
	"github.com/redis/go-redis/v9"
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
func generateOutputPath(outputDir string, o *models.Order) string {
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("order-%s-%s.pdf", o.OrderNumber, timestamp)
	return filepath.Join(outputDir, filename)
}

// A simple function for handling any potential redis streams ack errors. This is to reduce clutter at the original sites
func handleAckError(err error, logger *slog.Logger, m redis.XMessage) {
	if err != nil {
		logger.Error("An error occurred while acknowledging a stream message", "messageId", m.ID, "error", err)
	}
}
