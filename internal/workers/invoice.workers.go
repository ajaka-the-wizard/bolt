package workers

import (
	"context"
	"log/slog"
	"time"

	"github.com/ajaka-the-wizard/bolt/internal/domain"
	"github.com/ajaka-the-wizard/bolt/internal/store"
)

type invoiceWorkers struct {
	store *store.Store
}

func (i *invoiceWorkers) GenerateInvoice(ctx context.Context, id string) {
	logger := slog.Default().With(slog.String("group", domain.BoltRedisInvoiceConsumerGroup), slog.String("id", id))
	go func() {
		t := time.NewTicker(time.Second * 2)
		defer t.Stop()

		for range t.C {
			data, err := i.store.FetchNextTask(ctx, id, domain.BoltRedisInvoiceStreamKey, domain.BoltRedisInvoiceConsumerGroup, logger)
			if err == ctx.Err() {
				logger.Error("Cancellation error")
				break
			}
			logger.Info("Processing job", slog.Any("job", data))
			// TODO implement actual work
			time.Sleep(time.Minute)
		}

	}()
}

func InitInvoiceWorkers(ctx context.Context, store *store.Store, logger *slog.Logger) {
	s := invoiceWorkers{
		store: store,
	}
	s.GenerateInvoice(ctx, "worker1")
	s.GenerateInvoice(ctx, "worker2")
	s.GenerateInvoice(ctx, "worker3")
	logger.Info("Invoice generating workers initialized")
}
