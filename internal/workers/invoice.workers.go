package workers

import (
	"github.com/ajaka-the-wizard/bolt/internal/domain"
	"github.com/ajaka-the-wizard/bolt/internal/store"
)

type InvoiceWorkers struct {
	store *store.Store
}

func (i *InvoiceWorkers) GenerateInvoice(id string) {
	data := i.store.FetchNextTask(id, domain.BoltRedisInvoiceStreamKey, domain.BoltRedisInvoiceConsumerGroup)
	// TODO
}

func InitInvoiceWorkers() {

}
