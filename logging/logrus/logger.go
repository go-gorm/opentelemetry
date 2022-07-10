package logrus

import (
	"context"
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
)

type Writer struct {
	log *logrus.Logger
}

func NewWriter(opts ...Option) *Writer {
	cfg := defaultConfig()

	// apply options
	for _, opt := range opts {
		opt.apply(cfg)
	}

	// default trace hooks
	cfg.hooks = append(cfg.hooks, NewTraceHook(cfg.traceHookConfig))

	// attach hook
	for _, hook := range cfg.hooks {
		cfg.logger.AddHook(hook)
	}

	return &Writer{log: cfg.logger}
}

func (w *Writer) Logger() *logrus.Logger {
	return w.log
}

func (w *Writer) Trace(v ...interface{}) {
	w.log.Trace(v...)
}

func (w *Writer) Debug(v ...interface{}) {
	w.log.Debug(v...)
}

func (w *Writer) Info(v ...interface{}) {
	w.log.Info(v...)
}

func (w *Writer) Notice(v ...interface{}) {
	w.log.Warn(v...)
}

func (w *Writer) Warn(v ...interface{}) {
	w.log.Warn(v...)
}

func (w *Writer) Error(v ...interface{}) {
	w.log.Error(v...)
}

func (w *Writer) Fatal(v ...interface{}) {
	w.log.Fatal(v...)
}

func (w *Writer) Tracef(format string, v ...interface{}) {
	w.log.Tracef(format, v...)
}

func (w *Writer) Debugf(format string, v ...interface{}) {
	w.log.Debugf(format, v...)
}

func (w *Writer) Infof(format string, v ...interface{}) {
	w.log.Infof(format, v...)
}

func (w *Writer) Noticef(format string, v ...interface{}) {
	w.log.Warnf(format, v...)
}

func (w *Writer) Warnf(format string, v ...interface{}) {
	w.log.Warnf(format, v...)
}

func (w *Writer) Errorf(format string, v ...interface{}) {
	w.log.Errorf(format, v...)
}

func (w *Writer) Fatalf(format string, v ...interface{}) {
	w.log.Fatalf(format, v...)
}

func (w *Writer) CtxTracef(ctx context.Context, format string, v ...interface{}) {
	w.log.WithContext(ctx).Tracef(format, v...)
}

func (w *Writer) CtxDebugf(ctx context.Context, format string, v ...interface{}) {
	w.log.WithContext(ctx).Debugf(format, v...)
}

func (w *Writer) CtxInfof(ctx context.Context, format string, v ...interface{}) {
	w.log.WithContext(ctx).Infof(format, v...)
}

func (w *Writer) CtxNoticef(ctx context.Context, format string, v ...interface{}) {
	w.log.WithContext(ctx).Warnf(format, v...)
}

func (w *Writer) CtxWarnf(ctx context.Context, format string, v ...interface{}) {
	w.log.WithContext(ctx).Warnf(format, v...)
}

func (w *Writer) CtxErrorf(ctx context.Context, format string, v ...interface{}) {
	w.log.WithContext(ctx).Errorf(format, v...)
}

func (w *Writer) CtxFatalf(ctx context.Context, format string, v ...interface{}) {
	w.log.WithContext(ctx).Fatalf(format, v...)
}

func (w *Writer) SetOutput(writer io.Writer) {
	w.log.SetOutput(writer)
}

func (w *Writer) Printf(format string, v ...interface{}) {
	w.log.Info(fmt.Sprintf(format, v...))
}
