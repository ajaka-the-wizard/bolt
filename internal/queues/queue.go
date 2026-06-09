package queues

import "github.com/google/uuid"

type Queue struct{}

func InitQueue() *Queue {
	return &Queue{}
}

func (q *Queue) AddToReportGenQueue(id uuid.UUID) {
	// TODO Placeholder for queue
}
