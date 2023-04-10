package database

import (
	"context"

	"github.com/ice-coldbell/analyze-server/internal/model"
)

//go:generate go run github.com/golang/mock/mockgen -package=database -source=./database.go -destination=./mock_queue_test.go Database

type Database interface {
	Insert(context.Context, *model.Event) error
	Close() error
}
