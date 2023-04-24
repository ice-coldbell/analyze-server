package rabbitmq

import (
	"context"
	"encoding/json"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/ice-coldbell/analyze-server/pkg/errorx"
	"github.com/ice-coldbell/analyze-server/pkg/logger"
	amqp "github.com/rabbitmq/amqp091-go"
)

func New(cfg Config) (*rabbitMQ, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, errorx.Wrap(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, errorx.Wrap(err)
	}

	q, err := ch.QueueDeclare(
		cfg.QueueName, // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		return nil, errorx.Wrap(err)
	}

	return &rabbitMQ{
		conn:           conn,
		ch:             ch,
		q:              q,
		readerNum:      cfg.ReadLoop,
		handlerTimeout: time.Duration(cfg.HandlerTimeoutSec) * time.Second,
		handler:        make(map[string]func(context.Context) error),
		l:              logger.Root().Named("RABBITMQ"),
	}, nil
}

type rabbitMQ struct {
	conn      *amqp.Connection
	ch        *amqp.Channel
	q         amqp.Queue
	readerNum int

	handler map[string]func(context.Context) error
	l       logger.Logger

	readerWg       sync.WaitGroup
	handlerLock    sync.Mutex
	handlerTimeout time.Duration
}

func (q *rabbitMQ) Handle(message any, fn func(context.Context) error) {
	q.handlerLock.Lock()
	name := reflect.TypeOf(message).String()
	q.l.Debug("add handler function", logger.String("name", name))
	q.handler[name] = fn
	q.handlerLock.Unlock()
}

func (q *rabbitMQ) Enqueue(msg any) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return errorx.Wrap(err).With("message", msg)
	}
	if err := q.ch.PublishWithContext(
		context.Background(),
		"",
		q.q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        data,
			Type:        reflect.TypeOf(msg).String(),
		},
	); err != nil {
		return errorx.Wrap(err).With("message", msg)
	}
	return nil
}

func (q *rabbitMQ) ReadStart() {
	for i := 0; i < q.readerNum; i++ {
		q.readerWg.Add(1)
		loopNum := strconv.Itoa(i)
		go func() {
			if err := q.readLoop(loopNum); err != nil {
				q.l.WithError(err).Panic("start read loop")
			}
		}()
	}
}

func (q *rabbitMQ) Close() error {
	if !q.ch.IsClosed() {
		if err := q.ch.Close(); err != nil {
			return errorx.Wrap(err)
		}
	}

	if !q.conn.IsClosed() {
		if err := q.conn.Close(); err != nil {
			return errorx.Wrap(err)
		}
	}
	q.readerWg.Wait()
	return nil
}

func (q *rabbitMQ) readLoop(loopNum string) error {
	defer q.readerWg.Done()

	msgs, err := q.ch.Consume(q.q.Name, "", false, false, false, false, nil)
	if err != nil {
		return errorx.Wrap(err)
	}

	l := q.l.Named("READ").Named(loopNum)
	l.Debug("start read loop")

	for rmqMsg := range msgs {
		switch q.read(rmqMsg) {
		case nil:
			if err := rmqMsg.Ack(false); err != nil {
				l.WithError(errorx.Wrap(err)).Error("failed ack")
				continue
			}
		default:
			if err := rmqMsg.Nack(false, true); err != nil {
				l.WithError(errorx.Wrap(err)).
					With(
						logger.String("handler_name", rmqMsg.Type),
						logger.ByteString("data", rmqMsg.Body),
					).
					Error("failed ack")
				continue
			}
		}
	}
	return nil
}

func (q *rabbitMQ) read(rmqMsg amqp.Delivery) error {
	q.handlerLock.Lock()
	handle, ok := q.handler[rmqMsg.Type]
	q.handlerLock.Unlock()
	if !ok {
		return errorx.New("unknown message")
	}

	//lint:ignore SA1029 Only a single 'data' key is used.
	ctx := context.WithValue(context.Background(), "data", rmqMsg.Body)
	ctx, _ = context.WithTimeout(ctx, q.handlerTimeout)
	done := make(chan error)
	go func(ch chan error) {
		if err := handle(ctx); err != nil {
			ch <- err
			return
		}
		ch <- nil
	}(done)

	select {
	case err := <-done:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}
