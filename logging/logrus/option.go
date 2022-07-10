package logrus

import (
	"github.com/sirupsen/logrus"
)

type Option interface {
	apply(cfg *config)
}

type option func(cfg *config)

func (fn option) apply(cfg *config) {
	fn(cfg)
}

type config struct {
	logger *logrus.Logger
	hooks  []logrus.Hook

	traceHookConfig *TraceHookConfig
}

func defaultConfig() *config {
	// std logger
	stdLogger := logrus.StandardLogger()
	// default json format
	stdLogger.SetFormatter(new(logrus.JSONFormatter))

	return &config{
		logger: logrus.StandardLogger(),
		hooks:  []logrus.Hook{},
		traceHookConfig: &TraceHookConfig{
			recordStackTraceInSpan: true,
			enableLevels:           logrus.AllLevels,
			errorSpanLevel:         logrus.ErrorLevel,
		},
	}
}

func WithLogger(logger *logrus.Logger) Option {
	return option(func(cfg *config) {
		cfg.logger = logger
	})
}

func WithHook(hook logrus.Hook) Option {
	return option(func(cfg *config) {
		cfg.hooks = append(cfg.hooks, hook)
	})
}

func WithTraceHookConfig(hookConfig *TraceHookConfig) Option {
	return option(func(cfg *config) {
		cfg.traceHookConfig = hookConfig
	})
}

func WithTraceHookLevels(levels []logrus.Level) Option {
	return option(func(cfg *config) {
		cfg.traceHookConfig.enableLevels = levels
	})
}

func WithTraceHookErrorSpanLevel(level logrus.Level) Option {
	return option(func(cfg *config) {
		cfg.traceHookConfig.errorSpanLevel = level
	})
}

func WithRecordStackTraceInSpan(recordStackTraceInSpan bool) Option {
	return option(func(cfg *config) {
		cfg.traceHookConfig.recordStackTraceInSpan = true
	})
}
