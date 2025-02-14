package graceful

import (
	"sync"
	"testing"
)

func TestRunMultiple(t *testing.T) {
	g := newRoutineGroup()
	var mu sync.Mutex
	counter := 0
	iterations := 10

	for i := 0; i < iterations; i++ {
		g.Run(func() {
			mu.Lock()
			counter++
			mu.Unlock()
		})
	}

	g.Wait()

	if counter != iterations {
		t.Errorf("Expected counter to be %d, got %d", iterations, counter)
	}
}

func TestRunPanic(t *testing.T) {
	g := newRoutineGroup()
	panicRecovered := false
	done := make(chan struct{})

	g.Run(func() {
		defer func() {
			if r := recover(); r != nil {
				panicRecovered = true
			}
			close(done)
		}()
		panic("intentional panic for testing")
	})

	g.Wait()
	<-done

	if !panicRecovered {
		t.Error("Expected panic to be recovered")
	}
}
