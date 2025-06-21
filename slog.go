// Package graceful provides a Logger implementation using Go's log/slog.
package graceful

import (
	"fmt"
	"log/slog"
	"os"
)

// slogLogger implements Logger interface using log/slog.
type slogLogger struct {
	logger *slog.Logger
}

// SlogLoggerOption applies configuration to NewSlogLogger.
type SlogLoggerOption func(*slogLoggerOptions)

type slogLoggerOptions struct {
	logger *slog.Logger
	json   bool
}

// WithJSON returns an option to set output as JSON format.
func WithJSON() SlogLoggerOption {
	return func(opt *slogLoggerOptions) { opt.json = true }
}

// WithSlog injects a custom *slog.Logger instance.
func WithSlog(logger *slog.Logger) SlogLoggerOption {
	return func(opt *slogLoggerOptions) { opt.logger = logger }
}

// NewSlogLogger creates a Logger using flexible option pattern.
//
// Usage:
//
//	NewSlogLogger()                        // text mode (default)
//	NewSlogLogger(WithJson())              // json mode
//	NewSlogLogger(WithSlog(loggerObj))     // inject custom *slog.Logger, which overrides other options
func NewSlogLogger(opts ...SlogLoggerOption) Logger {
	var o slogLoggerOptions
	for _, f := range opts {
		f(&o)
	}
	if o.logger != nil {
		return &slogLogger{logger: o.logger}
	}
	var handler slog.Handler
	if o.json {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, nil)
	}
	return &slogLogger{
		logger: slog.New(handler),
	}
}

func (l *slogLogger) Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.logger.Info(msg)
}

func (l *slogLogger) Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.logger.Error(msg)
}
