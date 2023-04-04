package queue

import (
	"context"

	"github.com/ice-coldbell/analyze-server/pkg/errorx"
	"github.com/ice-coldbell/analyze-server/pkg/queue/kafka"
	"github.com/ice-coldbell/analyze-server/pkg/queue/rabbitmq"
	"gopkg.in/yaml.v3"
)

const (
	queueTypeRabbitMQ = "rabbitmq"
	queueTypeKafka    = "kafka"
)

type Queue interface {
	Handle(message any, fn func(context.Context) error)
	Enqueue(msg any) error
	Start()
	Stop() error
}

type Core struct {
	q         Queue `yaml:"-"`
	buildFunc func() error
}

func (cfg *Core) UnmarshalYAML(value *yaml.Node) error {
	var queueConfig struct {
		QueueType string `yaml:"type"`
	}
	if err := value.Decode(&queueConfig); err != nil {
		return errorx.Wrap(err)
	}
	switch queueConfig.QueueType {
	case queueTypeKafka:
		cfg.buildFunc = func() error {
			var kafkaConfig kafka.Config
			if err := value.Decode(&kafkaConfig); err != nil {
				return errorx.Wrap(err)
			}
			q, err := kafka.New(kafkaConfig)
			if err != nil {
				return err
			}
			cfg.q = q
			return nil
		}
	case queueTypeRabbitMQ:
		cfg.buildFunc = func() error {
			var rabbitmqConfig rabbitmq.Config
			if err := value.Decode(&rabbitmqConfig); err != nil {
				return errorx.Wrap(err)
			}
			q, err := rabbitmq.New(rabbitmqConfig)
			if err != nil {
				return err
			}
			cfg.q = q
			return nil
		}
	}
	return nil
}

func (cfg *Core) Build() error {
	if err := cfg.buildFunc(); err != nil {
		return err
	}
	cfg.buildFunc = nil
	return nil
}

func (cfg *Core) GetQueue() (Queue, error) {
	if cfg.q == nil {
		return nil, errorx.New("invaild queue type")
	}
	return cfg.q, nil
}
