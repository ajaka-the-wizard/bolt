package store

import (
	"context"
	"time"

	"github.com/ajaka-the-wizard/bolt/internal/models"
	"github.com/google/uuid"
)

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
