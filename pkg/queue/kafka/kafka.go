package kafka

import (
	"context"
	"encoding/json"
	"io"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/ice-coldbell/analyze-server/pkg/errorx"
	"github.com/ice-coldbell/analyze-server/pkg/logger"
	segmentioKafka "github.com/segmentio/kafka-go"
)

func New(cfg Config) (*core, error) {
	newCore := &core{
		handler: make(map[string]func(context.Context) error),
		l:       logger.Root().Named("KAFKA"),
	}

	if cfg.Reader != nil {
		newCore.reader = segmentioKafka.NewReader(segmentioKafka.ReaderConfig{
			GroupID: cfg.Reader.GroupID,
			Brokers: cfg.Reader.Brokers,
			Topic:   cfg.Reader.Topic,
		})
		newCore.handlerTimeout = time.Second * time.Duration(cfg.Reader.HandlerTimeoutSec)
		newCore.readerNum = cfg.Reader.ReadLoop
	}

	if cfg.Writer != nil {
		newCore.writer = &segmentioKafka.Writer{
			Addr:                   segmentioKafka.TCP(cfg.Writer.Brokers...),
			Topic:                  cfg.Writer.Topic,
			RequiredAcks:           cfg.Writer.RequiredAcks,
			Balancer:               &segmentioKafka.LeastBytes{},
			AllowAutoTopicCreation: false,
		}
	}

	return newCore, nil
}

type core struct {
	writer *segmentioKafka.Writer
	reader *segmentioKafka.Reader

	readerNum int
	readerWg  sync.WaitGroup

	handler        map[string]func(context.Context) error
	handlerLock    sync.Mutex
	handlerTimeout time.Duration

	l logger.Logger
}

func (c *core) Handle(message any, fn func(context.Context) error) {
	c.handlerLock.Lock()
	defer c.handlerLock.Unlock()

	handlerName := reflect.TypeOf(message).String()
	c.l.Debug("add handler function", logger.String("name", handlerName))
	c.handler[handlerName] = fn
}

func (c *core) Enqueue(message any) error {
	data, err := json.Marshal(message)
	if err != nil {
		return errorx.Wrap(err).With("message", message)
	}

	if err := c.writer.WriteMessages(
		context.Background(),
		segmentioKafka.Message{Key: []byte(reflect.TypeOf(message).String()), Value: data},
	); err != nil {
		return errorx.Wrap(err).With("message", message)
	}
	return nil
}

func (c *core) ReadStart() {
	for i := 0; i < c.readerNum; i++ {
		c.readerWg.Add(1)
		loopNum := strconv.Itoa(i) // copy i
		go c.readLoop(loopNum)
	}
}

func (c *core) Close() error {
	if c.writer != nil {
		if err := c.writer.Close(); err != nil {
			return errorx.Wrap(err)
		}
	}

	if c.reader != nil {
		if err := c.reader.Close(); err != nil {
			return errorx.Wrap(err)
		}
		c.readerWg.Wait()
	}
	return nil
}

func (c *core) readLoop(loopNum string) {
	defer c.readerWg.Done()

	l := c.l.Named("READ").Named(loopNum)
	l.Debug("start read loop")
	defer l.Debug("finish read loop")

	for {
		l := l
		kafkaMessage, err := c.reader.FetchMessage(context.Background())
		if err != nil {
			if !errorx.Is(err, io.EOF) {
				l.WithError(errorx.Wrap(err)).Error("fetch message")
			}
			break
		}

		handlerName := string(kafkaMessage.Key)
		l = l.With(
			logger.String("handler_name", handlerName),
			logger.ByteString("message", kafkaMessage.Value),
		)

		handle, ok := c.handler[handlerName]
		if !ok {
			l.Error("unknown message")
			continue
		}

		//lint:ignore SA1029 Only a single 'data' key is used.
		ctx := context.WithValue(context.Background(), "data", kafkaMessage.Value)
		ctx, cancel := context.WithTimeout(ctx, c.handlerTimeout)
		done := make(chan struct{})
		go func(ch chan struct{}) {
			if err := handle(ctx); err != nil {
				l.WithError(err).Error("unknown event")
				cancel()
				return
			}
			close(done)
		}(done)

		select {
		case <-done:
			// pass
		case <-ctx.Done():
			l.WithError(ctx.Err()).Error(ctx.Err().Error())
			continue
		}

		if err := c.reader.CommitMessages(ctx, kafkaMessage); err != nil {
			l.WithError(err).Error("commit message")
			continue
		}
	}
}
