package cassandra

import (
	"context"
	"time"

	"github.com/gocql/gocql"
	"github.com/ice-coldbell/analyze-server/internal/model"
	"github.com/ice-coldbell/analyze-server/pkg/errorx"
	"github.com/ice-coldbell/analyze-server/pkg/logger"
	"github.com/scylladb/gocqlx/v2"
)

func New(cfg Config) (*Database, error) {
	cluster := gocql.NewCluster(cfg.Hosts...)
	cluster.Keyspace = cfg.Keyspace
	cluster.ProtoVersion = 4
	session, err := gocqlx.WrapSession(cluster.CreateSession())
	if err != nil {
		return nil, errorx.Wrap(err)
	}

	return &Database{
		cluster: cluster,
		session: session,
		l:       logger.Root().Named("CASSANDRA"),
	}, nil
}

type Database struct {
	cluster *gocql.ClusterConfig
	session gocqlx.Session
	l       logger.Logger
}

func (db *Database) Insert(ctx context.Context, data *model.Event) error {
	batch := db.session.Session.NewBatch(gocql.LoggedBatch).WithContext(ctx)

	stmt, _ := tableEvent.Insert()
	batch.Query(stmt, data.ID, data.UserID, data.Identifier, data.EventTimestamp, data.Type)

	stmt, _ = tableEventData.Insert()
	batch.Query(stmt, data.ID, data.Data)

	stmt, _ = tableEventDate.Insert()
	eventDate := time.UnixMilli(data.EventTimestamp).Format(time.DateOnly)
	batch.Query(stmt, eventDate, data.EventTimestamp, data.ID)

	stmt, _ = tableEventUserID.Insert()
	batch.Query(stmt, data.UserID, data.Identifier, data.ID)

	if err := db.session.ExecuteBatch(batch); err != nil {
		return errorx.Wrap(err)
	}
	return nil
}

func (db *Database) Close() error {
	if !db.session.Closed() {
		db.session.Close()
	}
	return nil
}
