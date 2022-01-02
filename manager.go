package graceful

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Manager represents the graceful server manager interface
var manager *Manager

var initOnce = sync.Once{}

type RunningJob func(context.Context) error

// Manager manages the graceful shutdown process
type Manager struct {
	lock              *sync.RWMutex
	shutdownCtx       context.Context
	shutdownCtxCancel context.CancelFunc
	doneCtx           context.Context
	doneCtxCancel     context.CancelFunc
	logger            Logger
	runningWaitGroup  sync.WaitGroup
	errors            []error
}

func (g *Manager) start(ctx context.Context) {
	g.shutdownCtx, g.shutdownCtxCancel = context.WithCancel(ctx)
	g.doneCtx, g.doneCtxCancel = context.WithCancel(ctx)

	go g.handleSignals(ctx)
}

// DoGracefulShutdown graceful shutdown all task
func (g *Manager) DoGracefulShutdown() {
	g.shutdownCtxCancel()
	go func() {
		g.waitForJobs()
		g.lock.Lock()
		g.doneCtxCancel()
		g.lock.Unlock()
	}()
}

func (g *Manager) waitForJobs() {
	g.runningWaitGroup.Wait()
}

func (g *Manager) handleSignals(ctx context.Context) {
	c := make(chan os.Signal, 1)

	signal.Notify(
		c,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer signal.Stop(c)

	pid := syscall.Getpid()
	for {
		select {
		case sig := <-c:
			switch sig {
			case syscall.SIGINT:
				g.logger.Infof("PID %d. Received SIGINT. Shutting down...", pid)
				g.DoGracefulShutdown()
			case syscall.SIGTERM:
				g.logger.Infof("PID %d. Received SIGTERM. Shutting down...", pid)
				g.DoGracefulShutdown()
			default:
				g.logger.Infof("PID %d. Received %v.", pid, sig)
			}
		case <-ctx.Done():
			g.logger.Infof("PID: %d. Background context for manager closed - %v - Shutting down...", pid, ctx.Err())
			g.DoGracefulShutdown()
		}
	}
}

func (g *Manager) AddRunningJob(f RunningJob) {
	g.runningWaitGroup.Add(1)

	go func() {
		// to handle panic cases from inside the worker
		// in such case, we start a new goroutine
		defer func() {
			g.runningWaitGroup.Done()
			if err := recover(); err != nil {
				g.logger.Error(err)
			}
		}()
		if err := f(g.shutdownCtx); err != nil {
			g.lock.Lock()
			g.errors = append(g.errors, err)
			g.lock.Unlock()
		}
	}()
}

// Done allows the manager to be viewed as a context.Context.
func (g *Manager) Done() <-chan struct{} {
	return g.doneCtx.Done()
}

// NewManager initial the Manager
func NewManager(opts ...Option) *Manager {
	o := newOptions(opts...)
	initOnce.Do(func() {
		manager = &Manager{
			lock:   &sync.RWMutex{},
			logger: o.logger,
			errors: make([]error, 0),
		}
	})

	manager.start(o.ctx)

	return manager
}

// NewManager initial the Manager
func GetManager() *Manager {
	if manager == nil {
		panic("please new the manager first")
	}

	return manager
}
