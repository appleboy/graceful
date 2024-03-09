package graceful

// Logger interface is used throughout gorush
type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Info(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}

// NewEmptyLogger for simple logger.
func NewEmptyLogger() Logger {
	return emptyLogger{}
}

// EmptyLogger no meesgae logger
type emptyLogger struct{}

// Infof logs an informational message with a formatted string and optional arguments.
// It does not perform any logging in the emptyLogger implementation.
func (l emptyLogger) Infof(format string, args ...interface{})  {}
func (l emptyLogger) Errorf(format string, args ...interface{}) {}
func (l emptyLogger) Fatalf(format string, args ...interface{}) {}
func (l emptyLogger) Info(args ...interface{})                  {}
func (l emptyLogger) Error(args ...interface{})                 {}
func (l emptyLogger) Fatal(args ...interface{})                 {}
