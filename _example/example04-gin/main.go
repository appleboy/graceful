package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/appleboy/graceful"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.New()

	// Example route: GET /ping returns {"message": "pong"}
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	m := graceful.NewManager(
		graceful.WithLogger(graceful.NewSlogLogger(
			graceful.WithSlog(slog.Default()),
		)),
	)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Minute,
		WriteTimeout:      5 * time.Minute,
		MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
	}

	m.AddRunningJob(func(ctx context.Context) error {
		slog.Info("Starting HTTP server on :8080")
		return srv.ListenAndServe()
	})

	m.AddShutdownJob(func() error {
		slog.Info("Shutting down HTTP server gracefully")
		// Create a context with a timeout for the shutdown process
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return srv.Shutdown(ctx)
	})

	<-m.Done()
}
