package slog

import (
	"fmt"
	"io"
	"log/slog"
)

type Writer struct {
	log    *slog.Logger
	config *config
}

func NewWriter(opts ...Option) *Writer {
	cfg := defaultConfig()

	// apply options
	for _, opt := range opts {
		opt.apply(cfg)
	}
	// When user set the handlerOptions level but not set with coreconfig level
	if !cfg.coreConfig.withLevel && cfg.coreConfig.withHandlerOptions && cfg.coreConfig.opt.Level != nil {
		lvl := &slog.LevelVar{}
		lvl.Set(cfg.coreConfig.opt.Level.Level())
		cfg.coreConfig.level = lvl
	}
	cfg.coreConfig.opt.Level = cfg.coreConfig.level

	var replaceAttrDefined bool
	if cfg.coreConfig.opt.ReplaceAttr == nil {
		replaceAttrDefined = false
	} else {
		replaceAttrDefined = true
	}

	replaceFunc := cfg.coreConfig.opt.ReplaceAttr

	replaceAttr := func(groups []string, a slog.Attr) slog.Attr {
		// default replaceAttr level
		if a.Key == slog.LevelKey {
			level := a.Value.Any().(slog.Level)
			switch level {
			case slog.LevelDebug:
				a.Value = slog.StringValue("Debug")
			case slog.LevelInfo:
				a.Value = slog.StringValue("Info")
			case slog.LevelWarn:
				a.Value = slog.StringValue("Warn")
			case slog.LevelError:
				a.Value = slog.StringValue("Error")
			default:
				a.Value = slog.StringValue("Warn")
			}
		}
		// append replaceAttr by user
		if replaceAttrDefined {
			return replaceFunc(groups, a)
		} else {
			return a
		}
	}
	cfg.coreConfig.opt.ReplaceAttr = replaceAttr
	logger := slog.New(NewTraceHandler(cfg.coreConfig.writer, cfg.coreConfig.opt, cfg.traceConfig))

	return &Writer{log: logger, config: cfg}
}

func (w *Writer) Logger() *slog.Logger {
	return w.log
}

func (w *Writer) Printf(format string, v ...interface{}) {
	w.log.Info(fmt.Sprintf(format, v...))
}

func (w *Writer) SetLvel(level slog.Level) {
	w.config.coreConfig.level.Set(level)
	w.config.coreConfig.withLevel = true
}

func (w *Writer) SetOutput(writer io.Writer) {
	log := slog.New(NewTraceHandler(writer, w.config.coreConfig.opt, w.config.traceConfig))
	w.config.coreConfig.writer = writer
	w.log = log
}
