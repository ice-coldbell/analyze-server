package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Named(string) Logger
	With(...Field) Logger
	WithError(error) Logger

	Debug(string, ...Field)
	Info(string, ...Field)
	Warn(string, ...Field)
	Error(string, ...Field)
	Panic(string, ...Field)

	Shutdown() error
}

type TestLogger interface {
	Logger
	ObserverLogs() *ObservedLogs
}

func Root(opts ...BuildOption) Logger {
	buildRootLogger(opts...)
	return rootLogger
}

func RootTestLogger() TestLogger {
	buildRootLogger(withUseTestLogger())
	return rootLogger
}

type wrapLogger struct {
	l             *zap.Logger
	shutdownfuncs []func() error

	observerLogs *ObservedLogs //for test
}

func (l *wrapLogger) Debug(msg string, field ...Field) {
	l.l.Debug(msg, field...)
}

func (l *wrapLogger) Info(msg string, field ...Field) {
	l.l.Info(msg, field...)
}

func (l *wrapLogger) Warn(msg string, field ...Field) {
	l.l.Warn(msg, field...)
}

func (l *wrapLogger) Error(msg string, field ...Field) {
	l.l.Error(msg, field...)
}

func (l *wrapLogger) Panic(msg string, field ...Field) {
	l.l.Panic(msg, field...)
}

func (l *wrapLogger) Named(s string) Logger {
	return &wrapLogger{l: l.l.Named(s), observerLogs: l.observerLogs, shutdownfuncs: l.shutdownfuncs}
}

func (l *wrapLogger) With(fields ...Field) Logger {
	return &wrapLogger{l: l.l.With(fields...), observerLogs: l.observerLogs, shutdownfuncs: l.shutdownfuncs}
}

func (l *wrapLogger) WithError(err error) Logger {
	newlogger := &wrapLogger{observerLogs: l.observerLogs, shutdownfuncs: l.shutdownfuncs}
	switch terr := err.(type) {
	case zapcore.ObjectMarshaler:
		newlogger.l = l.l.With(Object("error", terr))
	case zapcore.ArrayMarshaler:
		newlogger.l = l.l.With(Array("errors", terr))
	default:
		newlogger.l = l.l.With(zap.Error(terr))
	}
	return newlogger
}

func (l *wrapLogger) ObserverLogs() *ObservedLogs {
	return l.observerLogs
}

func (l *wrapLogger) Shutdown() error {
	l.l.Sync()
	for i := range l.shutdownfuncs {
		if err := l.shutdownfuncs[i](); err != nil {
			return err
		}
	}
	return nil
}
