# graceful

[English](README.md) | **繁體中文** | [简体中文](README.zh-cn.md)

[![Run Tests](https://github.com/appleboy/graceful/actions/workflows/go.yml/badge.svg)](https://github.com/appleboy/graceful/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/appleboy/graceful/branch/master/graph/badge.svg?token=zPqtcz0Rum)](https://codecov.io/gh/appleboy/graceful)
[![Go Report Card](https://goreportcard.com/badge/github.com/appleboy/graceful)](https://goreportcard.com/report/github.com/appleboy/graceful)
[![Go Reference](https://pkg.go.dev/badge/github.com/gin-contrib/graceful.svg)](https://pkg.go.dev/github.com/gin-contrib/graceful)

一個輕量級 Go 語言套件，協助你優雅（graceful）地關閉服務與管理執行中的工作（job）。輕鬆管理長時間執行的任務與關閉時的清理 hook，確保你的服務能安全、可預期地結束。

---

## 目錄

- [graceful](#graceful)
  - [目錄](#目錄)
  - [功能特色](#功能特色)
  - [安裝方式](#安裝方式)
  - [使用說明](#使用說明)
    - [新增執行中工作](#新增執行中工作)
    - [新增關閉清理工作](#新增關閉清理工作)
    - [自訂 Logger](#自訂-logger)
  - [範例](#範例)
  - [授權](#授權)

---

## 功能特色

- 支援 Go 服務的優雅關閉（graceful shutdown）
- 管理多個執行中工作，並可透過 context 取消
- 註冊關閉時的清理 hook
- 支援自訂 logger
- API 簡單，易於整合

---

## 安裝方式

```bash
go get github.com/appleboy/graceful
```

---

## 使用說明

### 新增執行中工作

註冊長時間執行的工作，當服務關閉時會自動取消：

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

  // 新增工作 01
  m.AddRunningJob(func(ctx context.Context) error {
    for {
      select {
      case <-ctx.Done():
        return nil
      default:
        log.Println("執行工作 01")
        time.Sleep(1 * time.Second)
      }
    }
  })

  // 新增工作 02
  m.AddRunningJob(func(ctx context.Context) error {
    for {
      select {
      case <-ctx.Done():
        return nil
      default:
        log.Println("執行工作 02")
        time.Sleep(500 * time.Millisecond)
      }
    }
  })

  <-m.Done()
}
```

### 新增關閉清理工作

註冊關閉時要執行的清理邏輯：

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

  // 新增執行中工作（見上方範例）

  // 新增關閉清理 01
  m.AddShutdownJob(func() error {
    log.Println("關閉清理 01，等待 1 秒")
    time.Sleep(1 * time.Second)
    return nil
  })

  // 新增關閉清理 02
  m.AddShutdownJob(func() error {
    log.Println("關閉清理 02，等待 2 秒")
    time.Sleep(2 * time.Second)
    return nil
  })

  <-m.Done()
}
```

### 自訂 Logger

你可以使用自訂 logger（參考 [zerolog 範例](./_example/example03/logger.go)）：

```go
m := graceful.NewManager(
  graceful.WithLogger(logger{}),
)
```

---

## 範例

- [基本用法](./_example/example01/main.go)
- [多個工作](./_example/example02/main.go)
- [自訂 logger](./_example/example03/main.go)
- [Gin 整合](./_example/example04-gin/main.go)

---

## 授權

[MIT](LICENSE)
