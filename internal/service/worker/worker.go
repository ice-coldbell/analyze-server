package worker

import (
	"context"
	"encoding/json"

	"github.com/ice-coldbell/analyze-server/internal/infra/database"
	"github.com/ice-coldbell/analyze-server/internal/infra/queue"
	"github.com/ice-coldbell/analyze-server/internal/model"
	"github.com/ice-coldbell/analyze-server/pkg/errorx"
	"github.com/ice-coldbell/analyze-server/pkg/logger"
)

func New(cfg Config, q queue.Queue, db database.Database, l logger.Logger) *core {
	c := &core{
		db: db,
		q:  q,
		l:  l.Named("WORKER"),
	}

	q.Handle(model.Event{}, c.Handle())
	q.ReadStart()
	return c
}

type core struct {
	db database.Database
	q  queue.Queue

	l logger.Logger
}

func (c *core) Handle() func(context.Context) error {
	return func(ctx context.Context) error {
		ctxData, ok := ctx.Value("data").([]byte)
		if !ok {
			return errorx.New("invaild data type").With("data", ctx.Value("data"))
		}

		var event model.Event
		if err := json.Unmarshal(ctxData, &event); err != nil {
			return errorx.Wrap(err)
		}

		if err := c.db.Insert(ctx, &event); err != nil {
			return err
		}
		return nil
	}
}
