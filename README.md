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
    - [Add Running Jobs](#add-running-jobs)
    - [Add Shutdown Jobs](#add-shutdown-jobs)
    - [Custom Logger](#custom-logger)
  - [Examples](#examples)
  - [License](#license)

---

## Features

- Graceful shutdown for Go services
- Manage multiple running jobs with context cancellation
- Register shutdown hooks for cleanup tasks
- Custom logger support
- Simple API, easy integration

---

## Installation

```bash
go get github.com/appleboy/graceful
```

---

## Usage

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

### Custom Logger

You can use your own logger (see [zerolog example](./_example/example03/logger.go)):

```go
m := graceful.NewManager(
  graceful.WithLogger(logger{}),
)
```

---

## Examples

- [Basic usage](./_example/example01/main.go)
- [Multiple jobs](./_example/example02/main.go)
- [Custom logger](./_example/example03/main.go)
- [Gin integration](./_example/example04-gin/main.go)

---

## License

[MIT](LICENSE)
