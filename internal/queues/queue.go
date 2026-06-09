package queues

import "github.com/google/uuid"

type Queue struct{}

func InitQueue() *Queue {
	return &Queue{}
}

func (q *Queue) AddToReportGenQueue(id uuid.UUID) {
	// Placeholder for queue
	panic("Todo")
}
