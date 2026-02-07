package graceful

import (
	"context"
	"time"
)

// Option interface for configuration.
type Option interface {
	Apply(*Options)
}

// OptionFunc is a function that configures a graceful shutdown.
type OptionFunc func(*Options)

// Apply calls f(option)
func (f OptionFunc) Apply(option *Options) {
	f(option)
}

// Options for graceful shutdown
type Options struct {
	ctx             context.Context
	logger          Logger
	shutdownTimeout time.Duration
}

// WithContext custom context
func WithContext(ctx context.Context) Option {
	return OptionFunc(func(o *Options) {
		o.ctx = ctx
	})
}

// WithLogger custom logger
func WithLogger(logger Logger) Option {
	return OptionFunc(func(o *Options) {
		o.logger = logger
	})
}

// WithShutdownTimeout sets the maximum duration to wait for graceful shutdown to complete.
// If timeout is reached, the shutdown will proceed anyway and remaining jobs will be interrupted.
// A timeout of 0 means no timeout (wait indefinitely). Default is 30 seconds.
//
// Example:
//   m := graceful.NewManager(
//       graceful.WithShutdownTimeout(10 * time.Second),
//   )
func WithShutdownTimeout(timeout time.Duration) Option {
	return OptionFunc(func(o *Options) {
		o.shutdownTimeout = timeout
	})
}

// newOptions creates a new Options instance with default settings and applies any provided Option modifiers.
// It initializes the Options struct with a default background context and a new logger,
// then iterates over each given Option to adjust the configuration accordingly.
//
// Parameters:
//   - opts: A variadic list of Option functions that modify the default Options.
//
// Returns:
//   - Options: The customized Options struct after all modifications have been applied.
func newOptions(opts ...Option) Options {
	defaultOpts := Options{
		ctx:             context.Background(),
		logger:          NewLogger(),
		shutdownTimeout: 30 * time.Second,
	}

	// Loop through each option
	for _, opt := range opts {
		// Call the option giving the instantiated
		opt.Apply(&defaultOpts)
	}

	return defaultOpts
}
