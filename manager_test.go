package graceful

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func setup() {
	startOnce = sync.Once{}
}

func TestMissingManager(t *testing.T) {
	setup()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	_ = GetManager()
}

func TestManagerExist(t *testing.T) {
	setup()
	NewManager()
	m := GetManager()
	if m == nil {
		t.Errorf("missing manager")
	}
}

func TestRunningJob(t *testing.T) {
	setup()
	var count int32 = 0
	m := NewManager()

	// Add job
	m.AddRunningJob(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				atomic.AddInt32(&count, 1)
				time.Sleep(100 * time.Millisecond)
			}
		}
	})

	go func() {
		time.Sleep(50 * time.Millisecond)
		m.doGracefulShutdown()
	}()

	<-m.Done()

	if atomic.LoadInt32(&count) != 1 {
		t.Errorf("count error")
	}
}

func TestRunningAndShutdownJob(t *testing.T) {
	setup()
	var count int32 = 0
	m := NewManager()

	// Add job
	m.AddRunningJob(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				atomic.AddInt32(&count, 1)
				time.Sleep(100 * time.Millisecond)
			}
		}
	})

	m.AddShutdownJob(func() error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	go func() {
		time.Sleep(50 * time.Millisecond)
		m.doGracefulShutdown()
	}()

	<-m.Done()

	if atomic.LoadInt32(&count) != 2 {
		t.Errorf("count error: %v", atomic.LoadInt32(&count))
	}
}

func TestNewManagerWithContext(t *testing.T) {
	setup()
	ctx, cancel := context.WithCancel(context.Background())
	var count int32 = 0
	m := NewManagerWithContext(ctx)

	// Add job
	m.AddRunningJob(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				atomic.AddInt32(&count, 1)
				time.Sleep(100 * time.Millisecond)
			}
		}
	})

	m.AddShutdownJob(func() error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	<-m.Done()

	if atomic.LoadInt32(&count) != 2 {
		t.Errorf("count error: %v", atomic.LoadInt32(&count))
	}
}

func TestWithError(t *testing.T) {
	setup()
	ctx, cancel := context.WithCancel(context.Background())
	var count int32 = 0
	m := NewManagerWithContext(ctx)

	// Add job
	m.AddRunningJob(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				atomic.AddInt32(&count, 1)
				time.Sleep(100 * time.Millisecond)
				return errors.New("first error")
			}
		}
	})

	m.AddRunningJob(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				atomic.AddInt32(&count, 1)
				time.Sleep(100 * time.Millisecond)
				panic("four error")
			}
		}
	})

	m.AddShutdownJob(func() error {
		atomic.AddInt32(&count, 1)
		panic("second error")
	})

	m.AddShutdownJob(func() error {
		atomic.AddInt32(&count, 1)
		return errors.New("three error")
	})

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	<-m.Done()

	if atomic.LoadInt32(&count) != 4 {
		t.Errorf("count error: %v", atomic.LoadInt32(&count))
	}

	if len(m.errors) != 4 {
		t.Errorf("fail error count: %d", len(m.errors))
	}
}
