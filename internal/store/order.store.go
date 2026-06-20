package store

import (
	"context"
	"time"

	"github.com/ajaka-the-wizard/bolt/internal/models"
	"github.com/google/uuid"
)

// SaveOrder co-ordinates saving the order to the database, storing the idempotency key and adding a new invoice generation job to stream
func (s *Store) SaveOrder(ctx context.Context, data *models.Order, key string) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	id, err := s.db.SaveOrder(ctx, data)
	if err != nil {
		return uuid.UUID{}, err
	}
	if err := s.r.SetIdemKey(ctx, key); err != nil {
		return uuid.UUID{}, err
	}
	s.r.AddToInvoiceQueue(ctx, id)
	return id, nil
}

// FetchOrder fetches the order from database
func (s *Store) FetchOrder(ctx context.Context, id uuid.UUID, status models.Status, stage models.Stage) (*models.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.db.FetchOrder(ctx, id, status, stage)
}

func (s *Store) SetFailed(ctx context.Context, orderId uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.db.SetFailed(ctx, orderId)
}
