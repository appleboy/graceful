package graceful

import (
	"context"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
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
				time.Sleep(200 * time.Millisecond)
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
				time.Sleep(200 * time.Millisecond)
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

	// Test the public Errors() API
	errs := m.Errors()
	if len(errs) != 4 {
		t.Errorf("fail error count: %d", len(errs))
	}

	// Verify that returned slice is a copy (modifying it shouldn't affect internal state)
	errs[0] = errors.New("modified error")
	if m.Errors()[0].Error() == "modified error" {
		t.Errorf("Errors() should return a copy, not the internal slice")
	}
}

func TestGetShutdonwContext(t *testing.T) {
	setup()
	ctx, cancel := context.WithCancel(context.Background())
	var count int32 = 0
	m := NewManagerWithContext(ctx)

	m.AddShutdownJob(func() error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	m.AddShutdownJob(func() error {
		<-m.ShutdownContext().Done()
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

func TestWithSignalSIGINT(t *testing.T) {
	setup()
	testingSignal(t, syscall.SIGINT)
}

func TestWithSignalSIGTERM(t *testing.T) {
	setup()
	testingSignal(t, syscall.SIGTERM)
}

func testingSignal(t *testing.T, signal os.Signal) {
	var count int32 = 0
	m := NewManager()

	m.AddShutdownJob(func() error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	m.AddShutdownJob(func() error {
		<-m.ShutdownContext().Done()
		atomic.AddInt32(&count, 1)
		return nil
	})

	go func() {
		time.Sleep(50 * time.Millisecond)
		process, err := os.FindProcess(syscall.Getpid())
		if err != nil {
			t.Errorf("os.FindProcess error: %v", err)
		}
		if err := process.Signal(signal); err != nil {
			t.Errorf("process.Signal error: %v", err)
		}
	}()

	<-m.Done()

	if atomic.LoadInt32(&count) != 2 {
		t.Errorf("count error: %v", atomic.LoadInt32(&count))
	}
}

func TestShutdownTimeout(t *testing.T) {
	setup()
	var count int32 = 0
	m := NewManager(WithShutdownTimeout(100 * time.Millisecond))

	// Add a job that takes longer than timeout
	m.AddRunningJob(func(ctx context.Context) error {
		atomic.AddInt32(&count, 1)
		time.Sleep(500 * time.Millisecond) // This exceeds the 100ms timeout
		return nil
	})

	go func() {
		time.Sleep(50 * time.Millisecond)
		m.doGracefulShutdown()
	}()

	<-m.Done()

	// Job should have started
	if atomic.LoadInt32(&count) != 1 {
		t.Errorf("count error: %v", atomic.LoadInt32(&count))
	}

	// Should have timeout error
	errs := m.Errors()
	hasTimeoutError := false
	for _, err := range errs {
		if err.Error() == "shutdown timeout exceeded: 100ms" {
			hasTimeoutError = true
			break
		}
	}
	if !hasTimeoutError {
		t.Errorf("expected timeout error, got: %v", errs)
	}
}

func TestShutdownTimeoutZero(t *testing.T) {
	setup()
	var count int32 = 0
	m := NewManager(WithShutdownTimeout(0)) // No timeout

	// Add a job that takes a while
	m.AddRunningJob(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				atomic.AddInt32(&count, 1)
				return nil
			default:
				time.Sleep(50 * time.Millisecond)
			}
		}
	})

	go func() {
		time.Sleep(50 * time.Millisecond)
		m.doGracefulShutdown()
	}()

	<-m.Done()

	// Job should have completed
	if atomic.LoadInt32(&count) != 1 {
		t.Errorf("count error: %v", atomic.LoadInt32(&count))
	}

	// Should have no timeout error
	errs := m.Errors()
	for _, err := range errs {
		if err.Error() == "shutdown timeout exceeded: 0s" {
			t.Errorf("should not have timeout error with timeout=0")
		}
	}
}

func TestMultipleShutdownCalls(t *testing.T) {
	setup()
	var count int32 = 0
	m := NewManager()

	m.AddShutdownJob(func() error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	// Call shutdown multiple times
	go func() {
		time.Sleep(50 * time.Millisecond)
		for i := 0; i < 5; i++ {
			m.doGracefulShutdown()
			time.Sleep(10 * time.Millisecond)
		}
	}()

	<-m.Done()

	// Shutdown job should only run once despite multiple calls
	if atomic.LoadInt32(&count) != 1 {
		t.Errorf("shutdown job ran %d times, expected 1", atomic.LoadInt32(&count))
	}
}
