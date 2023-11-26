package graceful

import (
	"testing"
)

func TestRun(t *testing.T) {
	g := newRoutineGroup()

	// Define a flag to check if the function was executed
	var executed bool

	// Define the function to be executed
	fn := func() {
		executed = true
	}

	// Run the function in a goroutine
	g.Run(fn)

	// Wait for the goroutine to finish
	g.Wait()

	// Check if the function was executed
	if !executed {
		t.Error("Function was not executed")
	}
}
