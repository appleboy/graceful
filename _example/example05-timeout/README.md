# Example 05 - Shutdown Timeout

This example demonstrates the shutdown timeout feature added in the P1 improvements.

## Features Demonstrated

1. **Shutdown Timeout**: Configurable timeout to prevent indefinite waiting
2. **Multiple Shutdown Protection**: Repeated shutdown calls are handled gracefully
3. **Error Reporting**: Use `Errors()` to check for timeout or other errors

## Running the Example

```bash
cd _example/example05-timeout
go run main.go
```

Then press `Ctrl+C` to trigger graceful shutdown.

## What Happens

1. Two running jobs start:
   - **Fast job**: Completes within 1 second
   - **Slow job**: Takes too long (simulates stuck job)

2. When shutdown is triggered:
   - Both jobs receive context cancellation
   - Fast job exits immediately
   - Slow job takes 5+ seconds to exit
   - Timeout (3 seconds) is reached
   - Manager reports timeout error

3. Shutdown jobs run in parallel:
   - Database cleanup
   - Log flushing

## Expected Output

```txt
Fast job: received shutdown signal, exiting gracefully
Slow job: received shutdown signal, but taking too long...
Shutdown job: Closing database connections
Shutdown job: Flushing logs
ERROR: Shutdown timeout (3s) exceeded, some jobs may not have completed

⚠️  1 error(s) occurred during shutdown:
  1. shutdown timeout exceeded: 3s
```

## Configuration

Adjust the timeout using `WithShutdownTimeout`:

```go
// 10 second timeout
m := graceful.NewManager(
    graceful.WithShutdownTimeout(10 * time.Second),
)

// No timeout (wait forever)
m := graceful.NewManager(
    graceful.WithShutdownTimeout(0),
)
```

## Best Practices

1. Set a reasonable timeout based on your longest cleanup task
2. Always check `m.Errors()` after shutdown to detect timeouts
3. Design jobs to respond quickly to `ctx.Done()`
4. Consider Kubernetes termination grace period (default 30s)
