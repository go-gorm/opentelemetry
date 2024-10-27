package slog

import (
	"fmt"
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

	logger := slog.New(NewTraceHandler(cfg.coreConfig.writer, cfg.coreConfig.opt, cfg.traceConfig))

	return &Writer{log: logger, config: cfg}
}

func (w *Writer) Logger() *slog.Logger {
	return w.log
}

func (w *Writer) Printf(format string, v ...interface{}) {
	w.log.Info(fmt.Sprintf(format, v...))
}
