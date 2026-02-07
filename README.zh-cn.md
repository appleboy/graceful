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
    - [基本使用](#基本使用)
    - [新增运行中任务](#新增运行中任务)
    - [新增关闭清理任务](#新增关闭清理任务)
    - [设置关闭超时](#设置关闭超时)
    - [错误处理](#错误处理)
    - [自定义 Logger](#自定义-logger)
  - [配置选项](#配置选项)
  - [示例](#示例)
  - [最佳实践](#最佳实践)
    - [1. 务必等待 Done()](#1-务必等待-done)
    - [2. 响应 Context 取消](#2-响应-context-取消)
    - [3. 让关闭任务具有幂等性](#3-让关闭任务具有幂等性)
    - [4. 设置适当的超时](#4-设置适当的超时)
    - [5. 关闭后检查错误](#5-关闭后检查错误)
    - [6. 多个任务的关闭顺序](#6-多个任务的关闭顺序)
  - [许可证](#许可证)

---

## 功能特色

- **优雅关闭** - 自动处理系统信号（SIGINT、SIGTERM）的 Go 服务
- **超时保护** - 可配置超时时间防止无限期等待（默认：30 秒）
- **重复关闭保护** - 确保关闭逻辑只执行一次，即使收到多个信号
- **基于 Context 的取消** - 运行中任务会收到 context 取消信号
- **并行关闭 Hook** - 清理任务并行执行以加快关闭速度
- **错误报告** - 收集并报告所有任务的错误
- **自定义 Logger 支持** - 整合现有的日志解决方案
- **线程安全** - 所有操作都是并发安全的
- **零依赖** - 轻量且精简
- **简单的 API** - 易于集成到现有服务

---

## 安装方式

```bash
go get github.com/appleboy/graceful
```

---

## 使用说明

### 基本使用

创建 manager 并等待优雅关闭：

```go
package main

import (
  "context"
  "log"
  "time"

  "github.com/appleboy/graceful"
)

func main() {
  // 使用默认设置创建 manager
  m := graceful.NewManager()

  // 添加你的任务...

  // 等待关闭完成（会阻塞直到收到 SIGINT/SIGTERM）
  <-m.Done()

  // 检查错误
  if errs := m.Errors(); len(errs) > 0 {
    log.Printf("关闭过程中发生 %d 个错误", len(errs))
    for _, err := range errs {
      log.Printf("  - %v", err)
    }
  }

  log.Println("服务已优雅地停止")
}
```

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

### 设置关闭超时

设置等待优雅关闭的最长时间（默认：30 秒）：

```go
package main

import (
  "time"
  "github.com/appleboy/graceful"
)

func main() {
  // 设置 10 秒超时
  m := graceful.NewManager(
    graceful.WithShutdownTimeout(10 * time.Second),
  )

  // 或禁用超时（无限期等待）
  m := graceful.NewManager(
    graceful.WithShutdownTimeout(0),
  )

  // ... 添加任务 ...

  <-m.Done()

  // 检查是否发生超时
  if errs := m.Errors(); len(errs) > 0 {
    for _, err := range errs {
      if err.Error() == "shutdown timeout exceeded: 10s" {
        log.Println("部分任务未在超时内完成")
      }
    }
  }
}
```

**为什么超时很重要：**

- 防止任务未响应取消信号时无限期挂起
- 对容器化环境（Kubernetes terminationGracePeriodSeconds）至关重要
- 确保生产环境中可预测的关闭行为

### 错误处理

访问关闭期间发生的所有错误：

```go
package main

import (
  "log"
  "github.com/appleboy/graceful"
)

func main() {
  m := graceful.NewManager()

  m.AddRunningJob(func(ctx context.Context) error {
    // ... 执行任务 ...
    return fmt.Errorf("发生错误")  // 错误会被收集
  })

  m.AddShutdownJob(func() error {
    // ... 清理 ...
    return nil
  })

  <-m.Done()

  // 获取所有错误（包含任务错误、panic 和超时错误）
  errs := m.Errors()
  if len(errs) > 0 {
    log.Printf("关闭错误：%v", errs)
    os.Exit(1)  // 以错误代码退出
  }
}
```

**收集的错误类型：**

- 运行中任务返回的错误
- 关闭任务返回的错误
- 从任务中恢复的 panic（转换为错误）
- 如果关闭超过设置时间的超时错误

### 自定义 Logger

你可以使用自定义 logger（参考 [zerolog 示例](./_example/example03/logger.go)）：

```go
m := graceful.NewManager(
  graceful.WithLogger(logger{}),
)
```

---

## 配置选项

所有配置都通过传递给 `NewManager()` 的功能选项完成：

| 选项                            | 说明                                                    | 默认值                 |
| ------------------------------- | ------------------------------------------------------- | ---------------------- |
| `WithContext(ctx)`              | 使用自定义的父 context。当 context 被取消时会触发关闭。 | `context.Background()` |
| `WithLogger(logger)`            | 使用自定义的 logger 实现。                              | 内置 logger            |
| `WithShutdownTimeout(duration)` | 等待优雅关闭的最长时间。设为 `0` 表示无超时。           | `30 * time.Second`     |

**多个选项的示例：**

```go
m := graceful.NewManager(
  graceful.WithContext(ctx),
  graceful.WithShutdownTimeout(15 * time.Second),
  graceful.WithLogger(customLogger),
)
```

---

## 示例

- [**示例 01**：基本用法](./_example/example01/main.go) - 简单的运行中任务
- [**示例 02**：多个任务](./_example/example02/main.go) - 运行中 + 关闭任务
- [**示例 03**：自定义 logger](./_example/example03/main.go) - 与 zerolog 集成
- [**示例 04**：Gin 网页服务器](./_example/example04-gin/main.go) - 优雅的 HTTP 服务器关闭
- [**示例 05**：关闭超时](./_example/example05-timeout/main.go) - 超时设置与处理

---

## 最佳实践

### 1. 务必等待 Done()

```go
m := graceful.NewManager()
// ... 添加任务 ...
<-m.Done()  // ✅ 必要：等待关闭完成
```

**为什么：** 如果程序在调用 `<-m.Done()` 前就退出，清理可能无法完成，导致：

- 资源泄漏（打开的连接、文件）
- 数据丢失（未刷新的缓冲区）
- 孤儿 goroutine

### 2. 响应 Context 取消

```go
m.AddRunningJob(func(ctx context.Context) error {
  ticker := time.NewTicker(1 * time.Second)
  defer ticker.Stop()

  for {
    select {
    case <-ctx.Done():
      // ✅ 务必处理 ctx.Done() 以启用优雅关闭
      log.Println("正在优雅地关闭...")
      return ctx.Err()
    case <-ticker.C:
      // 执行任务
    }
  }
})
```

**为什么：** 不尊重 `ctx.Done()` 的任务会阻塞关闭直到超时。

### 3. 让关闭任务具有幂等性

```go
m.AddShutdownJob(func() error {
  // ✅ 多次调用也安全（虽然 graceful 确保只会调用一次）
  if db != nil {
    db.Close()
    db = nil
  }
  return nil
})
```

**为什么：** 虽然 manager 确保关闭任务只执行一次，防御性编程可防止问题。

### 4. 设置适当的超时

```go
// 对于 terminationGracePeriodSeconds: 30 的 Kubernetes Pod
m := graceful.NewManager(
  graceful.WithShutdownTimeout(25 * time.Second),  // ✅ 留 5 秒缓冲给 SIGKILL
)
```

**为什么：** 如果关闭超时超过容器终止期限，进程会被强制终止（SIGKILL）。

### 5. 关闭后检查错误

```go
<-m.Done()

if errs := m.Errors(); len(errs) > 0 {
  log.Printf("关闭错误：%v", errs)
  os.Exit(1)  // ✅ 以错误代码退出以便监控/告警
}
```

**为什么：** 让你能在生产环境中检测并响应关闭问题。

### 6. 多个任务的关闭顺序

关闭任务默认**并行执行**。如果你需要顺序关闭：

```go
m.AddShutdownJob(func() error {
  // 在单一任务内依序执行所有关闭步骤
  stopAcceptingRequests()
  waitForInflightRequests()
  closeDatabase()
  flushLogs()
  return nil
})
```

**为什么：** 并行执行速度较快，但某些清理需要特定顺序。

---

## 许可证

[MIT](LICENSE)
