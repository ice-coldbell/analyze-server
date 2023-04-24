package receiver

import (
	"github.com/ice-coldbell/analyze-server/core/infra/queue"
	"github.com/ice-coldbell/analyze-server/core/model"
	"github.com/ice-coldbell/analyze-server/pkg/logger"
)

func New(cfg Config, q queue.Queue, l logger.Logger) *core {
	c := &core{
		queue: q,
		l:     l.Named("RECEIVER"),

		stop: make(map[string]stopFunc),
	}

	if cfg.HTTP != nil && cfg.HTTP.Enable {
		c.httpReceiver(cfg.HTTP)
	}

	// if cfg.WebSocket != nil && cfg.WebSocket.Enable {

	// }

	// if cfg.GRPC != nil && cfg.GRPC.Enable {

	// }

	// if cfg.TCP != nil && cfg.TCP.Enable {

	// }

	q.ReadStart()
	return c
}

type stopFunc func() error

type core struct {
	queue queue.Queue
	l     logger.Logger

	stop map[string]stopFunc
}

func (c *core) Stop() {
	for key, stop := range c.stop {
		if err := stop(); err != nil {
			c.l.WithError(err).
				Error("fail stop receiver", logger.String("func_name", key))
		}
	}
}

func (c *core) enqueueEvent(event model.Event) {
	l := c.l.With(logger.Any("event", event))

	l.Debug("receive event")
	if err := c.queue.Enqueue(event); err != nil {
		l.WithError(err).Error("enqueue event")
		return
	}
	l.Debug("success enqueue event")
}

func (c *core) addStopFunction(name string, f stopFunc) {
	c.stop[name] = f
}
