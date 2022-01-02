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

// Manager manages the graceful shutdown process
type Manager struct {
	lock             *sync.RWMutex
	doneCtx          context.Context
	doneCtxCancel    context.CancelFunc
	logger           Logger
	runningWaitGroup sync.WaitGroup
	toRunTask        []func(context.Context) error
}

func (g *Manager) start(ctx context.Context) {
	g.doneCtx, g.doneCtxCancel = context.WithCancel(ctx)

	go g.handleSignals(ctx)
}

func (g *Manager) DoGracefulShutdown() {
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

// NewManager initial the Manager
func NewManager(opts ...Option) *Manager {
	o := newOptions(opts...)
	initOnce.Do(func() {
		manager = &Manager{
			lock:   &sync.RWMutex{},
			logger: o.logger,
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
