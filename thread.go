package graceful

import "sync"

// routineGroup represents a group of goroutines.
type routineGroup struct {
	waitGroup sync.WaitGroup
}

// newRoutineGroup creates a new routineGroup.
func newRoutineGroup() *routineGroup {
	return new(routineGroup)
}

// Run launches the given function fn in a new goroutine and tracks its execution using a wait group.
// It increments the wait group counter before starting the goroutine, and ensures the counter is decremented
// when fn completes, allowing for proper synchronization and cleanup.
func (g *routineGroup) Run(fn func()) {
	g.waitGroup.Add(1)

	go func() {
		defer g.waitGroup.Done()
		fn()
	}()
}

// Wait waits for all goroutines to finish.
func (g *routineGroup) Wait() {
	g.waitGroup.Wait()
}
