//go:build !go1.21

package graceful

import (
	"fmt"
	"log"
	"os"
)

// NewLogger for simple logger.
func NewLogger() Logger {
	return defaultLogger{
		infoLogger:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		fatalLogger: log.New(os.Stderr, "FATAL: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

type defaultLogger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	fatalLogger *log.Logger
}

func (l defaultLogger) Infof(format string, args ...interface{}) {
	l.infoLogger.Printf(format, args...)
}

func (l defaultLogger) Errorf(format string, args ...interface{}) {
	l.errorLogger.Printf(format, args...)
}

func (l defaultLogger) Fatalf(format string, args ...interface{}) {
	l.fatalLogger.Fatalf(format, args...)
}

func (l defaultLogger) Info(args ...interface{}) {
	l.infoLogger.Println(fmt.Sprint(args...))
}

func (l defaultLogger) Error(args ...interface{}) {
	l.errorLogger.Println(fmt.Sprint(args...))
}

func (l defaultLogger) Fatal(args ...interface{}) {
	l.fatalLogger.Println(fmt.Sprint(args...))
}
