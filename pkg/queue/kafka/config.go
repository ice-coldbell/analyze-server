package kafka

import (
	segmentioKafka "github.com/segmentio/kafka-go"
)

type Config struct {
	Reader *readerConfig `yaml:"reader"`
	Writer *writerConfig `yaml:"writer"`
}

type readerConfig struct {
	GroupID           string   `yaml:"groupID"`
	Brokers           []string `yaml:"brokers"`
	Topic             string   `yaml:"topic"`
	HandlerTimeoutSec int      `yaml:"timeout"` // Secound
	ReadLoop          int      `yaml:"readLoop"`
}

type writerConfig struct {
	Brokers      []string                    `yaml:"brokers"`
	Topic        string                      `yaml:"topic"`
	RequiredAcks segmentioKafka.RequiredAcks `yaml:"acks"`
}
