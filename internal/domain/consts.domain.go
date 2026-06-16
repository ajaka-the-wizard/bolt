package domain

import "path/filepath"

const (
	BoltRedisMaxRetries           = 5
	BoltIdempotencyKey            = "bolt:idempotency:"
	BoltRedisInvoiceStreamKey     = "bolt:queue:invoice:"
	BoltRedisWebhookStreamKey     = "bolt:queue:webhook:"
	BoltRedisEmailStreamKey       = "bolt:queue:email:"
	BoltRedisInvoiceConsumerGroup = "bolt:workers:group:invoice:"
	BoltRedisEmailConsumerGroup   = "bolt:workers:group:email:"
	BoltRedisWebhookConsumerGroup = "bolt:workers:group:webhook:"
	// We'll be storing the file to the server, it should be noted that it is more appropiate to store in an object storage
)

var (
	BoltInvoiceOutPutPath = filepath.Join("bolt", "invoice")
)
