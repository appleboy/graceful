package main

import (
	"context"
	"log"
	"time"

	"github.com/appleboy/graceful"
)

func main() {
	m := graceful.NewManager()

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

	<-m.Done()
}
