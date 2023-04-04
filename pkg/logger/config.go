package logger

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ice-coldbell/analyze-server/pkg/errorx"
	"github.com/ice-coldbell/lumberjack/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

// TODO: for test
var (
	filepathWalkDir = filepath.WalkDir
)

type config struct {
	Level        zap.AtomicLevel `yaml:"level"`
	Encoders     []Encoder       `yaml:"encoders"`
	ErrorOutputs []*Output       `yaml:"errorOutputs"`
}

type Encoder struct {
	Encoding string                `yaml:"encoding"`
	Outputs  []*Output             `yaml:"outputs"`
	Config   zapcore.EncoderConfig `yaml:"config"`
}

type Output struct {
	zapcore.WriteSyncer
	shutdownFunc func() error
}

func (out *Output) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		var scalar string
		if err := value.Decode(&scalar); err != nil {
			return errorx.Wrap(err)
		}
		switch scalar {
		case "stderr", "STDERR":
			out.WriteSyncer = zapcore.Lock(os.Stderr)
		case "stdout", "STDOUT":
			out.WriteSyncer = zapcore.Lock(os.Stdout)
		}
	case yaml.MappingNode:
		var logger lumberjack.Logger
		if err := value.Decode(&logger); err != nil {
			return errorx.Wrap(err)
		}
		out.shutdownFunc = func() error {
			loggerPtr := &logger
			if err := loggerPtr.Rotate(); err != nil {
				return errorx.Wrap(err)
			}
			loggerPtr.Shutdown()
			return os.Remove(loggerPtr.Filename)
		}
		out.WriteSyncer = zapcore.AddSync(&logger)
	}
	return nil
}

func getConfigFileName() string {
	const (
		defaultFileName       = "log.yaml"
		configFileNameEnvName = "LOG_CONFIG_FILE_NAME"
	)

	fileName := os.Getenv(configFileNameEnvName)
	if fileName == "" {
		fileName = defaultFileName
	}
	return fileName
}

func findConfigFilePath() (string, error) {
	fileName := getConfigFileName()

	var filePath string
	err := filepathWalkDir("/", func(path string, d fs.DirEntry, err error) error {
		if d.Name() == fileName {
			filePath = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", errorx.Wrap(err)
	}
	return filePath, nil
}

func loadConfig() (*config, error) {
	path, err := findConfigFilePath()
	if err != nil {
		return nil, errorx.Wrap(err)
	}
	data, err := os.Open(path)
	if err != nil {
		return nil, errorx.Wrap(err)
	}
	var cfg config
	err = yaml.NewDecoder(data).Decode(&cfg)
	if err != nil {
		return nil, errorx.Wrap(err)
	}
	return &cfg, nil
}
