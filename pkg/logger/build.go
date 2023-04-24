package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

var (
	rootLoggerOnce = sync.Once{}
	rootLogger     = &wrapLogger{}
)

type BuildOption interface {
	apply(*buildOption)
}

type buildOption struct {
	core       []zapcore.Core
	zapOptions []zap.Option
	isTest     bool
}

type optionFunc func(*buildOption)

func (f optionFunc) apply(opt *buildOption) {
	f(opt)
}

func WithCore(cores ...zapcore.Core) BuildOption {
	return optionFunc(func(opt *buildOption) { opt.core = cores })
}

func WithZapOption(options ...zap.Option) BuildOption {
	return optionFunc(func(opt *buildOption) { opt.zapOptions = options })
}

func withUseTestLogger() BuildOption {
	return optionFunc(func(opt *buildOption) {
		opt.isTest = true
	})
}

func buildRootLogger(opts ...BuildOption) *wrapLogger {
	rootLoggerOnce.Do(func() {
		cfg, err := loadConfig()
		if err != nil {
			panic(err)
		}

		var buildOpt buildOption
		for _, opt := range opts {
			opt.apply(&buildOpt)
		}

		if buildOpt.isTest {
			core, logs := observer.New(zap.InfoLevel)
			rootLogger.l = zap.New(core)
			rootLogger.observerLogs = logs
			return
		}

		// Add Shutdown Function
		var shutdownFuncs []func() error
		defer func() { rootLogger.shutdownfuncs = shutdownFuncs }()

		// Add Shutdown Function : lumberjack.Logger.Rotate
		for _, output := range cfg.ErrorOutputs {
			if output.shutdownFunc != nil {
				shutdownFuncs = append(shutdownFuncs, output.shutdownFunc)
			}
		}
		for _, enc := range cfg.Encoders {
			for _, output := range enc.Outputs {
				if output.shutdownFunc != nil {
					shutdownFuncs = append(shutdownFuncs, output.shutdownFunc)
				}
			}
		}

		// Build Core
		var cores []zapcore.Core
		for _, encCfg := range cfg.Encoders {
			var enc zapcore.Encoder
			switch encCfg.Encoding {
			case "json":
				enc = zapcore.NewJSONEncoder(encCfg.Config)
			default:
				enc = zapcore.NewConsoleEncoder(encCfg.Config)
			}
			var writeSyncer []zapcore.WriteSyncer
			for i := range encCfg.Outputs {
				writeSyncer = append(writeSyncer, encCfg.Outputs[i])
			}
			ws := zapcore.NewMultiWriteSyncer(writeSyncer...)
			cores = append(cores, zapcore.NewCore(enc, ws, cfg.Level))
		}

		// Add Option
		var options []zap.Option

		// Add Option : zap.ErrorOutput
		if len(cfg.ErrorOutputs) > 0 {
			var errorWriteSyncer []zapcore.WriteSyncer
			for i := range cfg.ErrorOutputs {
				errorWriteSyncer = append(errorWriteSyncer, cfg.ErrorOutputs[i])
			}
			errWS := zapcore.NewMultiWriteSyncer(errorWriteSyncer...)
			options = append(options, zap.ErrorOutput(errWS))
		}

		core := zapcore.NewTee(cores...)
		rootLogger.l = zap.New(core, options...)
	})
	return rootLogger
}
