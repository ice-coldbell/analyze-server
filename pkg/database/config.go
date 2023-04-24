package database

import (
	"context"

	"github.com/ice-coldbell/analyze-server/core/model"
	"github.com/ice-coldbell/analyze-server/pkg/database/cassandra"
	"github.com/ice-coldbell/analyze-server/pkg/errorx"
	"gopkg.in/yaml.v3"
)

const (
	databaseTypeCassandra = "cassandra"
)

type Database interface {
	Insert(context.Context, *model.Event) error
	Close() error
}

type Core struct {
	db        Database `yaml:"-"`
	buildFunc func() error
}

func (cfg *Core) UnmarshalYAML(value *yaml.Node) error {
	var DBConfig struct {
		DBType string `yaml:"type"`
	}
	if err := value.Decode(&DBConfig); err != nil {
		return errorx.Wrap(err)
	}
	switch t := DBConfig.DBType; t {
	case databaseTypeCassandra:
		cfg.buildFunc = func() error {
			var cassandraConfig cassandra.Config
			if err := value.Decode(&cassandraConfig); err != nil {
				return errorx.Wrap(err)
			}
			db, err := cassandra.New(cassandraConfig)
			if err != nil {
				return err
			}
			cfg.db = db
			return nil
		}
	default:
		return errorx.New("unknown database type").With("type", t)
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

func (cfg *Core) GetDatabase() (Database, error) {
	if cfg.db == nil {
		return nil, errorx.New("database is nil")
	}
	return cfg.db, nil
}
