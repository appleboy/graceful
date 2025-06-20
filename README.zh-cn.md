# graceful

[English](README.md) | [繁體中文](README.zh-tw.md) | **简体中文**

[![Run Tests](https://github.com/appleboy/graceful/actions/workflows/go.yml/badge.svg)](https://github.com/appleboy/graceful/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/appleboy/graceful/branch/master/graph/badge.svg?token=zPqtcz0Rum)](https://codecov.io/gh/appleboy/graceful)
[![Go Report Card](https://goreportcard.com/badge/github.com/appleboy/graceful)](https://goreportcard.com/report/github.com/appleboy/graceful)
[![Go Reference](https://pkg.go.dev/badge/github.com/gin-contrib/graceful.svg)](https://pkg.go.dev/github.com/gin-contrib/graceful)

一个轻量级的 Go 语言包，帮助你优雅（graceful）地关闭服务并管理运行中的任务（job）。轻松管理长时间运行的任务和关闭时的清理 hook，确保你的服务能够安全、可预期地退出。

---

## 目录

- [graceful](#graceful)
  - [目录](#目录)
  - [功能特色](#功能特色)
  - [安装方式](#安装方式)
  - [使用说明](#使用说明)
    - [新增运行中任务](#新增运行中任务)
    - [新增关闭清理任务](#新增关闭清理任务)
    - [自定义 Logger](#自定义-logger)
  - [示例](#示例)
  - [许可证](#许可证)

---

## 功能特色

- 支持 Go 服务的优雅关闭（graceful shutdown）
- 管理多个运行中任务，并可通过 context 取消
- 注册关闭时的清理 hook
- 支持自定义 logger
- API 简单，易于集成

---

## 安装方式

```bash
go get github.com/appleboy/graceful
```

---

## 使用说明

### 新增运行中任务

注册长时间运行的任务，服务关闭时会自动取消：

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

  // 新增任务 01
  m.AddRunningJob(func(ctx context.Context) error {
    for {
      select {
      case <-ctx.Done():
        return nil
      default:
        log.Println("执行任务 01")
        time.Sleep(1 * time.Second)
      }
    }
  })

  // 新增任务 02
  m.AddRunningJob(func(ctx context.Context) error {
    for {
      select {
      case <-ctx.Done():
        return nil
      default:
        log.Println("执行任务 02")
        time.Sleep(500 * time.Millisecond)
      }
    }
  })

  <-m.Done()
}
```

### 新增关闭清理任务

注册关闭时要执行的清理逻辑：

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

  // 新增运行中任务（见上方示例）

  // 新增关闭清理 01
  m.AddShutdownJob(func() error {
    log.Println("关闭清理 01，等待 1 秒")
    time.Sleep(1 * time.Second)
    return nil
  })

  // 新增关闭清理 02
  m.AddShutdownJob(func() error {
    log.Println("关闭清理 02，等待 2 秒")
    time.Sleep(2 * time.Second)
    return nil
  })

  <-m.Done()
}
```

### 自定义 Logger

你可以使用自定义 logger（参考 [zerolog 示例](./_example/example03/logger.go)）：

```go
m := graceful.NewManager(
  graceful.WithLogger(logger{}),
)
```

---

## 示例

- [基本用法](./_example/example01/main.go)
- [多个任务](./_example/example02/main.go)
- [自定义 logger](./_example/example03/main.go)
- [Gin 集成](./_example/example04-gin/main.go)

---

## 许可证

[MIT](LICENSE)
