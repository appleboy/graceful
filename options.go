package graceful

import "context"

// Option for queue system
type Option func(*options)

type options struct {
	ctx    context.Context
	logger Logger
}

func WithContext(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

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
