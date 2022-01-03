package main

import (
	"context"
	"log"
	"time"

	"github.com/appleboy/graceful"
)

func main() {
	m := graceful.NewManager(
		graceful.WithLogger(logger{}),
	)

	// Add job 01
	m.AddRunningJob(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				log.Println("working job 01")
				time.Sleep(1 * time.Second)
			}
		}
	})

	// Add job 02
	m.AddRunningJob(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				log.Println("working job 02")
				time.Sleep(500 * time.Millisecond)
			}
		}
	})

	// Add shutdown 01
	m.AddShutdownJob(func() error {
		log.Println("shutdown job 01 and wait 1 second")
		time.Sleep(1 * time.Second)
		return nil
	})

	// Add shutdown 02
	m.AddShutdownJob(func() error {
		log.Println("shutdown job 02 and wait 2 second")
		time.Sleep(2 * time.Second)
		return nil
	})

	<-m.Done()
}
