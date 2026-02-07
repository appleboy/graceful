package graceful

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// manager represents the graceful server manager interface
var manager *Manager

// startOnce initial graceful manager once
var startOnce = sync.Once{}

type (
	RunningJob  func(context.Context) error
	ShutdownJob func() error
)

// Manager manages the graceful shutdown process.
//
// The Manager uses a singleton pattern - only one instance can exist per process.
// It handles OS signals (SIGINT, SIGTERM) and context cancellation to trigger shutdown.
//
// Shutdown behavior:
//   - When a shutdown signal is received, all running jobs receive context cancellation
//   - Running jobs should respect context.Done() and exit gracefully
//   - After running jobs complete, shutdown jobs are executed in parallel
//   - If shutdown timeout is reached, remaining jobs are interrupted
//
// Signal handling:
//   - Unix: SIGINT (Ctrl+C), SIGTERM (kill), SIGTSTP
//   - Windows: SIGINT, SIGTERM
//   - Note: This will override any existing signal.Notify() for these signals
//
// Context behavior:
//   - If the parent context (from WithContext) is cancelled, shutdown is triggered
//   - ShutdownContext() returns a context that is cancelled when shutdown starts
//   - Done() returns a channel that is closed when all jobs complete
type Manager struct {
	lock              *sync.RWMutex
	shutdownCtx       context.Context
	shutdownCtxCancel context.CancelFunc
	doneCtx           context.Context
	doneCtxCancel     context.CancelFunc
	logger            Logger
	runningWaitGroup  *routineGroup
	errors            []error
	runAtShutdown     []ShutdownJob
	shutdownOnce      sync.Once
	shutdownTimeout   time.Duration
}

func (g *Manager) start(ctx context.Context) {
	g.shutdownCtx, g.shutdownCtxCancel = context.WithCancel(ctx)
	g.doneCtx, g.doneCtxCancel = context.WithCancel(context.Background())

	go g.handleSignals(ctx)
}

// doGracefulShutdown gracefully shuts down all tasks.
// This function is protected by sync.Once to ensure it only runs once,
// even if triggered multiple times by different signals or context cancellation.
func (g *Manager) doGracefulShutdown() {
	g.shutdownOnce.Do(func() {
		g.shutdownCtxCancel()

		// Copy shutdown jobs to avoid race condition while iterating
		g.lock.RLock()
		jobs := make([]ShutdownJob, len(g.runAtShutdown))
		copy(jobs, g.runAtShutdown)
		g.lock.RUnlock()

		// Execute shutdown jobs
		for _, f := range jobs {
			func(run ShutdownJob) {
				g.runningWaitGroup.Run(func() {
					g.doShutdownJob(run)
				})
			}(f)
		}

		go func() {
			g.waitForJobsWithTimeout()
			g.doneCtxCancel()
		}()
	})
}

func (g *Manager) waitForJobs() {
	g.runningWaitGroup.Wait()
}

// waitForJobsWithTimeout waits for all jobs to complete or until timeout is reached.
// If timeout is 0, it waits indefinitely.
func (g *Manager) waitForJobsWithTimeout() {
	if g.shutdownTimeout == 0 {
		// No timeout, wait indefinitely
		g.waitForJobs()
		return
	}

	done := make(chan struct{})
	go func() {
		g.waitForJobs()
		close(done)
	}()

	select {
	case <-done:
		// All jobs completed successfully
		g.logger.Infof("All jobs completed successfully")
	case <-time.After(g.shutdownTimeout):
		// Timeout reached
		g.logger.Errorf("Shutdown timeout (%v) exceeded, some jobs may not have completed", g.shutdownTimeout)
		msg := fmt.Errorf("shutdown timeout exceeded: %v", g.shutdownTimeout)
		g.lock.Lock()
		g.errors = append(g.errors, msg)
		g.lock.Unlock()
	}
}

func (g *Manager) handleSignals(ctx context.Context) {
	c := make(chan os.Signal, 1)

	signal.Notify(
		c,
		signals...,
	)
	defer signal.Stop(c)

	pid := syscall.Getpid()
	for {
		select {
		case sig := <-c:
			switch sig {
			case syscall.SIGINT:
				g.logger.Infof("PID %d. Received SIGINT. Shutting down...", pid)
				g.doGracefulShutdown()
				return
			case syscall.SIGTERM:
				g.logger.Infof("PID %d. Received SIGTERM. Shutting down...", pid)
				g.doGracefulShutdown()
				return
			default:
				g.logger.Infof("PID %d. Received %v.", pid, sig)
			}
		case <-ctx.Done():
			g.logger.Infof("PID: %d. Background context for manager closed - %v - Shutting down...", pid, ctx.Err())
			g.doGracefulShutdown()
			return
		}
	}
}

// doShutdownJob execute shutdown task
func (g *Manager) doShutdownJob(f ShutdownJob) {
	// to handle panic cases from inside the worker
	defer func() {
		if err := recover(); err != nil {
			msg := fmt.Errorf("panic in shutdown job: %v", err)
			g.logger.Errorf(msg.Error())
			g.lock.Lock()
			g.errors = append(g.errors, msg)
			g.lock.Unlock()
		}
	}()
	if err := f(); err != nil {
		g.lock.Lock()
		g.errors = append(g.errors, err)
		g.lock.Unlock()
	}
}

// AddShutdownJob adds a shutdown task that will be executed when graceful shutdown is triggered.
//
// Shutdown jobs are executed in parallel after all running jobs have completed.
// Each shutdown job should be idempotent as they may be called during unexpected shutdowns.
//
// Note: This method is thread-safe and can be called from multiple goroutines.
// However, jobs added after shutdown has started will not be executed.
func (g *Manager) AddShutdownJob(f ShutdownJob) {
	g.lock.Lock()
	g.runAtShutdown = append(g.runAtShutdown, f)
	g.lock.Unlock()
}

// AddRunningJob adds a long-running task that will receive shutdown signals via context.
//
// Running jobs should:
//   - Monitor ctx.Done() and exit gracefully when signaled
//   - Return an error if something goes wrong
//   - Clean up resources before returning
//
// Example:
//   m.AddRunningJob(func(ctx context.Context) error {
//       ticker := time.NewTicker(1 * time.Second)
//       defer ticker.Stop()
//       for {
//           select {
//           case <-ctx.Done():
//               return nil  // Graceful exit
//           case <-ticker.C:
//               // Do work
//           }
//       }
//   })
//
// Note: This method is thread-safe. Panics are recovered and converted to errors.
func (g *Manager) AddRunningJob(f RunningJob) {
	g.runningWaitGroup.Run(func() {
		// to handle panic cases from inside the worker
		defer func() {
			if err := recover(); err != nil {
				msg := fmt.Errorf("panic in running job: %v", err)
				g.logger.Errorf(msg.Error())
				g.lock.Lock()
				g.errors = append(g.errors, msg)
				g.lock.Unlock()
			}
		}()
		if err := f(g.shutdownCtx); err != nil {
			g.lock.Lock()
			g.errors = append(g.errors, err)
			g.lock.Unlock()
		}
	})
}

// Done returns a channel that is closed when all jobs (running + shutdown) have completed.
//
// This should be used to block the main goroutine until graceful shutdown is complete:
//
//   m := graceful.NewManager()
//   // ... add jobs ...
//   <-m.Done()  // Block until shutdown completes
//   errs := m.Errors()  // Check for errors
//
// Warning: If you don't wait for Done(), your program may exit before cleanup completes,
// potentially causing goroutine leaks or incomplete shutdown.
func (g *Manager) Done() <-chan struct{} {
	return g.doneCtx.Done()
}

// ShutdownContext returns a context that is cancelled when shutdown begins.
//
// Use this context for operations that should be cancelled during shutdown.
// This is the same context passed to running jobs.
//
// Example:
//   ctx := m.ShutdownContext()
//   req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
//   // Request will be cancelled when shutdown starts
func (g *Manager) ShutdownContext() context.Context {
	return g.shutdownCtx
}

// Errors returns all errors that occurred during running jobs and shutdown jobs.
//
// This includes:
//   - Errors returned by running jobs
//   - Errors returned by shutdown jobs
//   - Panics recovered from jobs (converted to errors)
//   - Timeout errors if shutdown exceeded the configured timeout
//
// The returned slice is a copy, so modifying it won't affect the internal state.
//
// Example:
//   <-m.Done()
//   if errs := m.Errors(); len(errs) > 0 {
//       for _, err := range errs {
//           log.Printf("Shutdown error: %v", err)
//       }
//       os.Exit(1)
//   }
func (g *Manager) Errors() []error {
	g.lock.RLock()
	defer g.lock.RUnlock()
	// Return a copy to prevent external modification
	result := make([]error, len(g.errors))
	copy(result, g.errors)
	return result
}

func newManager(opts ...Option) *Manager {
	startOnce.Do(func() {
		o := newOptions(opts...)
		manager = &Manager{
			lock:             &sync.RWMutex{},
			logger:           o.logger,
			errors:           make([]error, 0),
			runningWaitGroup: newRoutineGroup(),
			shutdownTimeout:  o.shutdownTimeout,
		}
		manager.start(o.ctx)
	})

	return manager
}

// NewManager creates and initializes the graceful shutdown Manager.
//
// This function uses a singleton pattern - calling it multiple times returns the same instance.
// The Manager automatically starts listening for OS signals (SIGINT, SIGTERM) and will trigger
// graceful shutdown when received.
//
// Options:
//   - WithContext(ctx): Use a custom parent context. Shutdown triggers when context is cancelled.
//   - WithLogger(logger): Use a custom logger implementation.
//   - WithShutdownTimeout(duration): Set maximum time to wait for shutdown (default: 30s).
//
// Example:
//   m := graceful.NewManager(
//       graceful.WithShutdownTimeout(10 * time.Second),
//       graceful.WithLogger(customLogger),
//   )
//
// Important: Only one Manager can exist per process.
func NewManager(opts ...Option) *Manager {
	return newManager(opts...)
}

// NewManagerWithContext initial the Manager with custom context
func NewManagerWithContext(ctx context.Context, opts ...Option) *Manager {
	return newManager(append(opts, WithContext(ctx))...)
}

// GetManager returns the existing Manager instance.
//
// This will panic if NewManager() has not been called first.
// Use this in places where you need access to the manager but can't pass it directly.
//
// Example:
//   // In main.go
//   m := graceful.NewManager()
//
//   // In another package
//   m := graceful.GetManager()
//   m.AddShutdownJob(cleanup)
func GetManager() *Manager {
	if manager == nil {
		panic("please use NewManager to initial the manager first")
	}

	return manager
}
