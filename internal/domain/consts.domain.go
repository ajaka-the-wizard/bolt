package domain

const (
	BoltIdempotencyKey            = "bolt:idempotency:"
	BoltRedisInvoiceStreamKey     = "bolt:queue:invoice:"
	BoltRedisWebhookStreamKey     = "bolt:queue:webhook:"
	BoltRedisEmailStreamKey       = "bolt:queue:email:"
	BoltRedisInvoiceConsumerGroup = "bolt:workers:group:invoice:"
	BoltRedisEmailConsumerGroup   = "bolt:workers:group:email:"
	BoltRedisWebhookConsumerGroup = "bolt:workers:group:webhook:"
)
