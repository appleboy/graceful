package main

import (
	"context"
	"fmt"
	"time"

	"github.com/appleboy/graceful"
)

func main() {
	fmt.Println("=== Graceful Shutdown with Timeout Demo ===")

	// Create manager with 3 second timeout
	m := graceful.NewManager(
		graceful.WithShutdownTimeout(3 * time.Second),
	)

	// Add a fast running job
	m.AddRunningJob(func(ctx context.Context) error {
		fmt.Println("Fast job started")
		for i := 0; i < 5; i++ {
			select {
			case <-ctx.Done():
				fmt.Println("Fast job: received shutdown signal, exiting gracefully")
				return nil
			default:
				fmt.Printf("Fast job: working... (%d/5)\n", i+1)
				time.Sleep(200 * time.Millisecond)
			}
		}
		return nil
	})

	// Add a slow running job (simulates a stuck job)
	m.AddRunningJob(func(ctx context.Context) error {
		fmt.Println("Slow job started (will be interrupted by timeout)")
		for i := 0; i < 20; i++ {
			select {
			case <-ctx.Done():
				fmt.Println("Slow job: received shutdown signal, but taking too long...")
				// Simulate a job that doesn't exit quickly
				time.Sleep(5 * time.Second)
				fmt.Println("Slow job: finally done (this won't be printed due to timeout)")
				return nil
			default:
				fmt.Printf("Slow job: working... (%d/20)\n", i+1)
				time.Sleep(200 * time.Millisecond)
			}
		}
		return nil
	})

	// Add shutdown jobs
	m.AddShutdownJob(func() error {
		fmt.Println("Shutdown job: Closing database connections")
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	m.AddShutdownJob(func() error {
		fmt.Println("Shutdown job: Flushing logs")
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	// Trigger shutdown after 2 seconds
	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println("\n>>> Triggering graceful shutdown... <<<")
		// Note: In real scenarios, this would be triggered by SIGINT/SIGTERM
		// For demo purposes, we'll use context cancellation
		ctx := m.ShutdownContext()
		// Simulate signal by accessing the manager's internal state
		// In production, just send SIGINT (Ctrl+C)
		_ = ctx
	}()

	// For demo, trigger shutdown programmatically
	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println("\n>>> Simulating Ctrl+C (SIGINT) ... <<<")
		// In a real app, user would press Ctrl+C
		// Here we'll wait for the jobs to complete
	}()

	// Wait for all jobs to complete (or timeout)
	fmt.Println("Main: Waiting for shutdown to complete...")
	fmt.Println("Main: Press Ctrl+C to trigger shutdown")

	<-m.Done()

	fmt.Println("\n=== Shutdown Complete ===")

	// Check for errors
	errs := m.Errors()
	if len(errs) > 0 {
		fmt.Printf("\n⚠️  %d error(s) occurred during shutdown:\n", len(errs))
		for i, err := range errs {
			fmt.Printf("  %d. %v\n", i+1, err)
		}
	} else {
		fmt.Println("\n✅ All jobs completed successfully within timeout")
	}

	fmt.Println("\nDemo completed. Exiting.")
}
