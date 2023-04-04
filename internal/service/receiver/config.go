package receiver

type Config struct {
	HTTP      *httpConfig      `yaml:"http"`
	WebSocket *websocketConfig `yaml:"websocket"`
	GRPC      *grpcConfig      `yaml:"grpc"`
	TCP       *tcpConfig       `yaml:"tcp"`
}

type httpConfig struct {
	Path               string `yaml:"path"`
	Port               string `yaml:"port"`
	ShutdownTimeoutSec int    `yaml:"shutdownTimeoutSec"`
	Enable             bool   `yaml:"enable"`
}

type websocketConfig struct {
	Path   string `yaml:"path"`
	Enable bool   `yaml:"enable"`
}

type grpcConfig struct {
	Enable bool `yaml:"enable"`
}

type tcpConfig struct {
	Enable bool `yaml:"enable"`
}
