package queue

import (
	"context"
)

//go:generate go run github.com/golang/mock/mockgen -package=queue -source=./queue.go -destination=./mock_queue_test.go Queue,QueueMessage,QueueMessageID

type Queue interface {
	Handle(message any, fn func(context.Context) error)
	Enqueue(message any) error
	ReadStart()
	Close() error
}
