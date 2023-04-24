package rabbitmq

type Config struct {
	URL               string `yaml:"url"`
	QueueName         string `yaml:"name"`
	HandlerTimeoutSec int    `yaml:"timeout"` // Secound
	ReadLoop          int    `yaml:"readLoop"`
}
