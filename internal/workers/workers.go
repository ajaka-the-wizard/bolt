package workers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ajaka-the-wizard/bolt/internal/domain"
	"github.com/ajaka-the-wizard/bolt/internal/errs"
	"github.com/ajaka-the-wizard/bolt/internal/models"
	"github.com/ajaka-the-wizard/bolt/internal/store"
	"github.com/google/uuid"
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

// Simple determinitic pdf file name generator, could be improved? maybe.
func generateOutputPath(outputDir string, o *models.Order) string {
	timestamp := o.CreatedAt.Format("20060102-150405")
	filename := fmt.Sprintf("order-%s-%s.pdf", o.OrderNumber, timestamp)
	return filepath.Join(outputDir, filename)
}

// Helper to ensure no invoice is processed twice
func tryGetPath(outputDir string, o *models.Order) (string, bool) {
	path := generateOutputPath(outputDir, o)
	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		return "", false
	}
	return path, true
}

// A simple function for handling any potential redis streams ack errors. This is to reduce clutter at the original sites
func handleAckError(err error, logger *slog.Logger, m redis.XMessage) {
	if err != nil {
		logger.Error("An error occurred while acknowledging a stream message", "messageId", m.ID, "error", err)
	}
}

// A reusable helper for handling retry errors. It also helps keep the main function lean and readable
func handleRetriesError(ctx context.Context, logger *slog.Logger, orderId uuid.UUID, s *store.Store) {
	logger.Warn("Received data with exhausted retries, setting as failed in db", "orderId", orderId)
	if err := s.SetFailed(ctx, orderId); err != nil {
		if errors.Is(err, errs.ErrOrderNoExists) {
			logger.Warn("Order doesnt exist", "orderID", orderId)
			return
		}
		logger.Error("Could not set failed", "error", err, "orderId", orderId)
		return
	}
	logger.Info("Setting order as failed succeded", "orderId", orderId)
}

// A reusable helper for handling order fetching errors. It also helps keep the main function lean and readable
func handleFetchError(ctx context.Context, err error, logger *slog.Logger, orderId uuid.UUID, store *store.Store, m redis.XMessage) {
	if errors.Is(err, errs.ErrOrderNoExists) {
		logger.Warn("Order does not exists, dropping message", "order_id", orderId)
		err = store.Ack(ctx, domain.BoltRedisInvoiceStreamKey, domain.BoltRedisInvoiceConsumerGroup, m.ID)
		handleAckError(err, logger, m)
		return
	}
	logger.Error("Failed to fetch order", "error", err, "order_id", orderId)
}

// ParseRedisValue parses my redis stream values, returns them all including a bool to signify if they are all valid or not.
func parseRedisValue(v map[string]any) (uuid.UUID, int, int, bool) {

	id, ok := v["order_id"].(string)
	cId, err := uuid.Parse(id)
	maxRetries, err := strconv.Atoi(v["max_retries"].(string))
	noOfRetries, err := strconv.Atoi(v["no_of_retries"].(string))
	if err != nil {
		ok = false
	}
	return cId, maxRetries, noOfRetries, ok
}
