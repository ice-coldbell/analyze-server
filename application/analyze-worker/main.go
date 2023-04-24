package main

import (
	"os"
	"os/signal"

	"github.com/ice-coldbell/analyze-server/core/config"
	"github.com/ice-coldbell/analyze-server/core/service/worker"
	"github.com/ice-coldbell/analyze-server/pkg/errorx"
	"github.com/ice-coldbell/analyze-server/pkg/logger"
)

func main() {
	l := logger.Root().Named("WORKER")
	defer func() {
		if err := l.Shutdown(); err != nil {
			l.WithError(errorx.Wrap(err)).Error("failed logger shutdown")
			return
		}
	}()

	var cfg config.WorkerConfig
	if err := config.LoadConfig(&cfg); err != nil {
		l.WithError(errorx.Wrap(err)).Error("failed load config")
		return
	}

	if err := cfg.Queue.Build(); err != nil {
		l.WithError(errorx.Wrap(err)).Error("failed build queue")
		return
	}

	if err := cfg.DB.Build(); err != nil {
		l.WithError(errorx.Wrap(err)).Error("failed build database")
		return
	}

	eventQueue, err := cfg.Queue.GetQueue()
	if err != nil {
		l.WithError(err).Error("failed get queue")
		return
	}
	defer func() {
		if err := eventQueue.Close(); err != nil {
			l.WithError(err).Error("failed queue shutdown")
		}
	}()

	eventDB, err := cfg.DB.GetDatabase()
	if err != nil {
		l.WithError(err).Error("failed get database")
		return
	}
	defer func() {
		if err := eventDB.Close(); err != nil {
			l.WithError(err).Error("failed database shutdown")
		}
	}()

	worker.New(cfg.Worker, eventQueue, eventDB, l)

	l.Debug("RUNNING...")
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown
	l.Debug("SHUTDOWN")
}
