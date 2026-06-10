package domain

var (
	BoltIdempotencyKey     = "bolt:idempotency:"
	BoltRedisStreamKey     = "bolt:jobs:key"
	BoltRedisConsumerGroup = "bolt:workers:group"
)
