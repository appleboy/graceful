# graceful

**English** | [繁體中文](README.zh-tw.md) | [简体中文](README.zh-cn.md)

[![Run Tests](https://github.com/appleboy/graceful/actions/workflows/go.yml/badge.svg)](https://github.com/appleboy/graceful/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/appleboy/graceful/branch/master/graph/badge.svg?token=zPqtcz0Rum)](https://codecov.io/gh/appleboy/graceful)
[![Go Report Card](https://goreportcard.com/badge/github.com/appleboy/graceful)](https://goreportcard.com/report/github.com/appleboy/graceful)
[![Go Reference](https://pkg.go.dev/badge/github.com/gin-contrib/graceful.svg)](https://pkg.go.dev/github.com/gin-contrib/graceful)

A lightweight Go package for graceful shutdown and job management. Easily manage long-running jobs and shutdown hooks, ensuring your services exit cleanly and predictably.

---

## Table of Contents

- [graceful](#graceful)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Installation](#installation)
  - [Usage](#usage)
    - [Basic Usage](#basic-usage)
    - [Add Running Jobs](#add-running-jobs)
    - [Add Shutdown Jobs](#add-shutdown-jobs)
    - [Configure Shutdown Timeout](#configure-shutdown-timeout)
    - [Error Handling](#error-handling)
    - [Custom Logger](#custom-logger)
  - [Configuration Options](#configuration-options)
  - [Examples](#examples)
  - [Best Practices](#best-practices)
    - [1. Always Wait for Done()](#1-always-wait-for-done)
    - [2. Respond to Context Cancellation](#2-respond-to-context-cancellation)
    - [3. Make Shutdown Jobs Idempotent](#3-make-shutdown-jobs-idempotent)
    - [4. Set Appropriate Timeout](#4-set-appropriate-timeout)
    - [5. Check Errors After Shutdown](#5-check-errors-after-shutdown)
    - [6. Shutdown Order with Multiple Jobs](#6-shutdown-order-with-multiple-jobs)
  - [License](#license)

---

## Features

- **Graceful shutdown** for Go services with automatic signal handling (SIGINT, SIGTERM)
- **Timeout protection** - configurable timeout to prevent indefinite hanging (default: 30s)
- **Multiple shutdown protection** - ensures shutdown logic only runs once, even with multiple signals
- **Context-based cancellation** - running jobs receive context cancellation signals
- **Parallel shutdown hooks** - cleanup tasks run concurrently for faster shutdown
- **Error reporting** - collect and report all errors from jobs
- **Custom logger support** - integrate with your existing logging solution
- **Thread-safe** - all operations are safe for concurrent use
- **Zero dependencies** - lightweight and minimal
- **Simple API** - easy integration with existing services

---

## Installation

```bash
go get github.com/appleboy/graceful
```

---

## Usage

### Basic Usage

Create a manager and wait for graceful shutdown:

```go
package main

import (
  "context"
  "log"
  "time"

  "github.com/appleboy/graceful"
)

func main() {
  // Create a manager with default settings
  m := graceful.NewManager()

  // Add your jobs...

  // Wait for shutdown to complete (blocks until SIGINT/SIGTERM received)
  <-m.Done()

  // Check for errors
  if errs := m.Errors(); len(errs) > 0 {
    log.Printf("Shutdown completed with %d error(s)", len(errs))
    for _, err := range errs {
      log.Printf("  - %v", err)
    }
  }

  log.Println("Service stopped gracefully")
}
```

### Add Running Jobs

Register long-running jobs that will be cancelled on shutdown:

```go
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
```

### Add Shutdown Jobs

Register shutdown hooks to run cleanup logic before exit:

```go
package main

import (
  "context"
  "log"
  "time"

  "github.com/appleboy/graceful"
)

func main() {
  m := graceful.NewManager()

  // Add running jobs (see above)

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
```

### Configure Shutdown Timeout

Set a maximum time to wait for graceful shutdown (default: 30 seconds):

```go
package main

import (
  "time"
  "github.com/appleboy/graceful"
)

func main() {
  // Set 10 second timeout
  m := graceful.NewManager(
    graceful.WithShutdownTimeout(10 * time.Second),
  )

  // Or disable timeout (wait indefinitely)
  m := graceful.NewManager(
    graceful.WithShutdownTimeout(0),
  )

  // ... add jobs ...

  <-m.Done()

  // Check if timeout occurred
  if errs := m.Errors(); len(errs) > 0 {
    for _, err := range errs {
      if err.Error() == "shutdown timeout exceeded: 10s" {
        log.Println("Some jobs did not complete within timeout")
      }
    }
  }
}
```

**Why timeout matters:**

- Prevents indefinite hanging if a job doesn't respond to cancellation
- Critical for containerized environments (Kubernetes terminationGracePeriodSeconds)
- Ensures predictable shutdown behavior in production

### Error Handling

Access all errors that occurred during shutdown:

```go
package main

import (
  "log"
  "github.com/appleboy/graceful"
)

func main() {
  m := graceful.NewManager()

  m.AddRunningJob(func(ctx context.Context) error {
    // ... do work ...
    return fmt.Errorf("something went wrong")  // Error will be collected
  })

  m.AddShutdownJob(func() error {
    // ... cleanup ...
    return nil
  })

  <-m.Done()

  // Get all errors (includes job errors, panics, and timeout errors)
  errs := m.Errors()
  if len(errs) > 0 {
    log.Printf("Shutdown errors: %v", errs)
    os.Exit(1)  // Exit with error code
  }
}
```

**Error types collected:**

- Errors returned by running jobs
- Errors returned by shutdown jobs
- Panics recovered from jobs (converted to errors)
- Timeout errors if shutdown exceeds configured duration

### Custom Logger

You can use your own logger (see [zerolog example](./_example/example03/logger.go)):

```go
m := graceful.NewManager(
  graceful.WithLogger(logger{}),
)
```

---

## Configuration Options

All configuration is done through functional options passed to `NewManager()`:

| Option                          | Description                                                               | Default                |
| ------------------------------- | ------------------------------------------------------------------------- | ---------------------- |
| `WithContext(ctx)`              | Use a custom parent context. Shutdown triggers when context is cancelled. | `context.Background()` |
| `WithLogger(logger)`            | Use a custom logger implementation.                                       | Built-in logger        |
| `WithShutdownTimeout(duration)` | Maximum time to wait for graceful shutdown. Set to `0` for no timeout.    | `30 * time.Second`     |

**Example with multiple options:**

```go
m := graceful.NewManager(
  graceful.WithContext(ctx),
  graceful.WithShutdownTimeout(15 * time.Second),
  graceful.WithLogger(customLogger),
)
```

---

## Examples

- [**Example 01**: Basic usage](./_example/example01/main.go) - Simple running jobs
- [**Example 02**: Multiple jobs](./_example/example02/main.go) - Running + shutdown jobs
- [**Example 03**: Custom logger](./_example/example03/main.go) - Integration with zerolog
- [**Example 04**: Gin web server](./_example/example04-gin/main.go) - Graceful HTTP server shutdown
- [**Example 05**: Shutdown timeout](./_example/example05-timeout/main.go) - Timeout configuration and handling

---

## Best Practices

### 1. Always Wait for Done()

```go
m := graceful.NewManager()
// ... add jobs ...
<-m.Done()  // ✅ REQUIRED: Wait for shutdown to complete
```

**Why:** If your program exits before calling `<-m.Done()`, cleanup may not complete, leading to:

- Resource leaks (open connections, files)
- Data loss (unflushed buffers)
- Orphaned goroutines

### 2. Respond to Context Cancellation

```go
m.AddRunningJob(func(ctx context.Context) error {
  ticker := time.NewTicker(1 * time.Second)
  defer ticker.Stop()

  for {
    select {
    case <-ctx.Done():
      // ✅ Always handle ctx.Done() to enable graceful shutdown
      log.Println("Shutting down gracefully...")
      return ctx.Err()
    case <-ticker.C:
      // Do work
    }
  }
})
```

**Why:** Jobs that don't respect `ctx.Done()` will block shutdown until timeout is reached.

### 3. Make Shutdown Jobs Idempotent

```go
m.AddShutdownJob(func() error {
  // ✅ Safe to call multiple times (though graceful ensures it's only called once)
  if db != nil {
    db.Close()
    db = nil
  }
  return nil
})
```

**Why:** Although the manager ensures shutdown jobs only run once, defensive coding prevents issues.

### 4. Set Appropriate Timeout

```go
// For Kubernetes pods with terminationGracePeriodSeconds: 30
m := graceful.NewManager(
  graceful.WithShutdownTimeout(25 * time.Second),  // ✅ Leave 5s buffer for SIGKILL
)
```

**Why:** If your shutdown timeout exceeds the container termination period, the process will be forcefully killed (SIGKILL).

### 5. Check Errors After Shutdown

```go
<-m.Done()

if errs := m.Errors(); len(errs) > 0 {
  log.Printf("Shutdown errors: %v", errs)
  os.Exit(1)  // ✅ Exit with error code for monitoring/alerting
}
```

**Why:** Allows you to detect and respond to shutdown issues in production.

### 6. Shutdown Order with Multiple Jobs

Shutdown jobs run in **parallel** by design. If you need sequential shutdown:

```go
m.AddShutdownJob(func() error {
  // Do all shutdown in sequence within a single job
  stopAcceptingRequests()
  waitForInflightRequests()
  closeDatabase()
  flushLogs()
  return nil
})
```

**Why:** Parallel execution is faster, but some cleanup requires specific ordering.

---

## License

[MIT](LICENSE)
