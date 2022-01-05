package graceful

import "context"

// Option for Functional
type Option func(*options)

type options struct {
	ctx    context.Context
	logger Logger
}

// WithContext custom context
func WithContext(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

// WithLogger custom logger
func WithLogger(logger Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

func newOptions(opts ...Option) options {
	defaultOpts := options{
		ctx:    context.Background(),
		logger: NewLogger(),
	}

	// Loop through each option
	for _, opt := range opts {
		// Call the option giving the instantiated
		opt(&defaultOpts)
	}

	return defaultOpts
}
