package main

import (
	"os"
	"os/signal"

	"github.com/ice-coldbell/analyze-server/internal/config"
	"github.com/ice-coldbell/analyze-server/internal/service/receiver"
	"github.com/ice-coldbell/analyze-server/pkg/errorx"
	"github.com/ice-coldbell/analyze-server/pkg/logger"
)

func main() {
	l := logger.Root().Named("SERVER")
	defer func() {
		if err := l.Shutdown(); err != nil {
			l.WithError(errorx.Wrap(err)).Error("failed logger shutdown")
			return
		}
	}()

	var cfg config.ServerConfig
	if err := config.LoadConfig(&cfg); err != nil {
		l.WithError(errorx.Wrap(err)).Error("failed load config")
		return
	}

	if err := cfg.Queue.Build(); err != nil {
		l.WithError(errorx.Wrap(err)).Error("failed build queue")
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

	eventReceiver := receiver.New(cfg.Receiver, eventQueue, l)
	defer eventReceiver.Stop()

	l.Debug("RUNNING...")
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown
	l.Debug("SHUTDOWN")
}
