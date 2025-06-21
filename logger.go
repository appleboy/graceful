package graceful

import (
	"log"
	"os"
)

// Logger interface is used throughout gorush
type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// NewLogger for simple logger.
func NewLogger() Logger {
	return defaultLogger{
		infoLogger:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

type defaultLogger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
}

func (l defaultLogger) Infof(format string, args ...interface{}) {
	l.infoLogger.Printf(format, args...)
}

func (l defaultLogger) Errorf(format string, args ...interface{}) {
	l.errorLogger.Printf(format, args...)
}

// NewEmptyLogger for simple logger.
func NewEmptyLogger() Logger {
	return emptyLogger{}
}

// EmptyLogger no meesgae logger
type emptyLogger struct{}

func (l emptyLogger) Infof(format string, args ...interface{})  {}
func (l emptyLogger) Errorf(format string, args ...interface{}) {}
