package config

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ice-coldbell/analyze-server/internal/service/receiver"
	"github.com/ice-coldbell/analyze-server/internal/service/worker"
	"github.com/ice-coldbell/analyze-server/pkg/database"
	"github.com/ice-coldbell/analyze-server/pkg/errorx"
	"github.com/ice-coldbell/analyze-server/pkg/queue"
	"gopkg.in/yaml.v3"
)

// TODO: for test
var (
	filepathWalkDir = filepath.WalkDir
	osOpen          = os.Open
)

type ServerConfig struct {
	Receiver receiver.Config `yaml:"receiver"`
	Queue    queue.Core      `yaml:"queue"`
}

func (c ServerConfig) FileName() string {
	return "server.yaml"
}

type WorkerConfig struct {
	Worker worker.Config `yaml:"worker"`
	Queue  queue.Core    `yaml:"queue"`
	DB     database.Core `yaml:"db`
}

func (c WorkerConfig) FileName() string {
	return "worker.yaml"
}

type IConfig interface {
	FileName() string
}

func LoadConfig(cfg IConfig) error {
	path, err := findConfigFilePath(cfg.FileName())
	if err != nil {
		return err
	}
	data, err := osOpen(path)
	if err != nil {
		return errorx.Wrap(err)
	}
	err = yaml.NewDecoder(data).Decode(cfg)
	if err != nil {
		return err
	}
	return nil
}

func findConfigFilePath(fileName string) (string, error) {
	var filePath string
	err := filepathWalkDir("/", func(path string, d fs.DirEntry, err error) error {
		if d.Name() == fileName {
			filePath = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return filePath, nil
}
