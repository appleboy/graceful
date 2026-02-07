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
    - [基本使用](#基本使用)
    - [新增執行中工作](#新增執行中工作)
    - [新增關閉清理工作](#新增關閉清理工作)
    - [設定關閉逾時](#設定關閉逾時)
    - [錯誤處理](#錯誤處理)
    - [自訂 Logger](#自訂-logger)
  - [設定選項](#設定選項)
  - [範例](#範例)
  - [最佳實踐](#最佳實踐)
    - [1. 務必等待 Done()](#1-務必等待-done)
    - [2. 回應 Context 取消](#2-回應-context-取消)
    - [3. 讓關閉工作具有冪等性](#3-讓關閉工作具有冪等性)
    - [4. 設定適當的逾時](#4-設定適當的逾時)
    - [5. 關閉後檢查錯誤](#5-關閉後檢查錯誤)
    - [6. 多個工作的關閉順序](#6-多個工作的關閉順序)
  - [授權](#授權)

---

## 功能特色

- **優雅關閉** - 自動處理系統信號（SIGINT、SIGTERM）的 Go 服務
- **逾時保護** - 可設定逾時時間防止無限期等待（預設：30 秒）
- **重複關閉保護** - 確保關閉邏輯只執行一次，即使收到多個信號
- **基於 Context 的取消** - 執行中工作會收到 context 取消信號
- **並行關閉 Hook** - 清理任務並行執行以加快關閉速度
- **錯誤回報** - 收集並回報所有工作的錯誤
- **自訂 Logger 支援** - 整合現有的日誌解決方案
- **執行緒安全** - 所有操作都是並行安全的
- **零相依性** - 輕量且精簡
- **簡單的 API** - 易於整合到現有服務

---

## 安裝方式

```bash
go get github.com/appleboy/graceful
```

---

## 使用說明

### 基本使用

建立 manager 並等待優雅關閉：

```go
package main

import (
  "context"
  "log"
  "time"

  "github.com/appleboy/graceful"
)

func main() {
  // 使用預設設定建立 manager
  m := graceful.NewManager()

  // 新增你的工作...

  // 等待關閉完成（會阻塞直到收到 SIGINT/SIGTERM）
  <-m.Done()

  // 檢查錯誤
  if errs := m.Errors(); len(errs) > 0 {
    log.Printf("關閉過程中發生 %d 個錯誤", len(errs))
    for _, err := range errs {
      log.Printf("  - %v", err)
    }
  }

  log.Println("服務已優雅地停止")
}
```

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

### 設定關閉逾時

設定等待優雅關閉的最長時間（預設：30 秒）：

```go
package main

import (
  "time"
  "github.com/appleboy/graceful"
)

func main() {
  // 設定 10 秒逾時
  m := graceful.NewManager(
    graceful.WithShutdownTimeout(10 * time.Second),
  )

  // 或停用逾時（無限期等待）
  m := graceful.NewManager(
    graceful.WithShutdownTimeout(0),
  )

  // ... 新增工作 ...

  <-m.Done()

  // 檢查是否發生逾時
  if errs := m.Errors(); len(errs) > 0 {
    for _, err := range errs {
      if err.Error() == "shutdown timeout exceeded: 10s" {
        log.Println("部分工作未在逾時內完成")
      }
    }
  }
}
```

**為什麼逾時很重要：**

- 防止工作未回應取消信號時無限期掛起
- 對容器化環境（Kubernetes terminationGracePeriodSeconds）至關重要
- 確保生產環境中可預測的關閉行為

### 錯誤處理

存取關閉期間發生的所有錯誤：

```go
package main

import (
  "log"
  "github.com/appleboy/graceful"
)

func main() {
  m := graceful.NewManager()

  m.AddRunningJob(func(ctx context.Context) error {
    // ... 執行工作 ...
    return fmt.Errorf("發生錯誤")  // 錯誤會被收集
  })

  m.AddShutdownJob(func() error {
    // ... 清理 ...
    return nil
  })

  <-m.Done()

  // 取得所有錯誤（包含工作錯誤、panic 和逾時錯誤）
  errs := m.Errors()
  if len(errs) > 0 {
    log.Printf("關閉錯誤：%v", errs)
    os.Exit(1)  // 以錯誤代碼退出
  }
}
```

**收集的錯誤類型：**

- 執行中工作回傳的錯誤
- 關閉工作回傳的錯誤
- 從工作中復原的 panic（轉換為錯誤）
- 如果關閉超過設定時間的逾時錯誤

### 自訂 Logger

你可以使用自訂 logger（參考 [zerolog 範例](./_example/example03/logger.go)）：

```go
m := graceful.NewManager(
  graceful.WithLogger(logger{}),
)
```

---

## 設定選項

所有設定都透過傳遞給 `NewManager()` 的功能選項完成：

| 選項                            | 說明                                                  | 預設值                 |
| ------------------------------- | ----------------------------------------------------- | ---------------------- |
| `WithContext(ctx)`              | 使用自訂的父 context。當 context 被取消時會觸發關閉。 | `context.Background()` |
| `WithLogger(logger)`            | 使用自訂的 logger 實作。                              | 內建 logger            |
| `WithShutdownTimeout(duration)` | 等待優雅關閉的最長時間。設為 `0` 表示無逾時。         | `30 * time.Second`     |

**多個選項的範例：**

```go
m := graceful.NewManager(
  graceful.WithContext(ctx),
  graceful.WithShutdownTimeout(15 * time.Second),
  graceful.WithLogger(customLogger),
)
```

---

## 範例

- [**範例 01**：基本用法](./_example/example01/main.go) - 簡單的執行中工作
- [**範例 02**：多個工作](./_example/example02/main.go) - 執行中 + 關閉工作
- [**範例 03**：自訂 logger](./_example/example03/main.go) - 與 zerolog 整合
- [**範例 04**：Gin 網頁伺服器](./_example/example04-gin/main.go) - 優雅的 HTTP 伺服器關閉
- [**範例 05**：關閉逾時](./_example/example05-timeout/main.go) - 逾時設定與處理

---

## 最佳實踐

### 1. 務必等待 Done()

```go
m := graceful.NewManager()
// ... 新增工作 ...
<-m.Done()  // ✅ 必要：等待關閉完成
```

**為什麼：** 如果程式在呼叫 `<-m.Done()` 前就退出，清理可能無法完成，導致：

- 資源洩漏（開放的連線、檔案）
- 資料遺失（未刷新的緩衝區）
- 孤兒 goroutine

### 2. 回應 Context 取消

```go
m.AddRunningJob(func(ctx context.Context) error {
  ticker := time.NewTicker(1 * time.Second)
  defer ticker.Stop()

  for {
    select {
    case <-ctx.Done():
      // ✅ 務必處理 ctx.Done() 以啟用優雅關閉
      log.Println("正在優雅地關閉...")
      return ctx.Err()
    case <-ticker.C:
      // 執行工作
    }
  }
})
```

**為什麼：** 不尊重 `ctx.Done()` 的工作會阻塞關閉直到逾時。

### 3. 讓關閉工作具有冪等性

```go
m.AddShutdownJob(func() error {
  // ✅ 多次呼叫也安全（雖然 graceful 確保只會呼叫一次）
  if db != nil {
    db.Close()
    db = nil
  }
  return nil
})
```

**為什麼：** 雖然 manager 確保關閉工作只執行一次，防禦性編程可防止問題。

### 4. 設定適當的逾時

```go
// 對於 terminationGracePeriodSeconds: 30 的 Kubernetes Pod
m := graceful.NewManager(
  graceful.WithShutdownTimeout(25 * time.Second),  // ✅ 留 5 秒緩衝給 SIGKILL
)
```

**為什麼：** 如果關閉逾時超過容器終止期限，程序會被強制終止（SIGKILL）。

### 5. 關閉後檢查錯誤

```go
<-m.Done()

if errs := m.Errors(); len(errs) > 0 {
  log.Printf("關閉錯誤：%v", errs)
  os.Exit(1)  // ✅ 以錯誤代碼退出以便監控/告警
}
```

**為什麼：** 讓你能在生產環境中偵測並回應關閉問題。

### 6. 多個工作的關閉順序

關閉工作預設**並行執行**。如果你需要循序關閉：

```go
m.AddShutdownJob(func() error {
  // 在單一工作內依序執行所有關閉步驟
  stopAcceptingRequests()
  waitForInflightRequests()
  closeDatabase()
  flushLogs()
  return nil
})
```

**為什麼：** 並行執行速度較快，但某些清理需要特定順序。

---

## 授權

[MIT](LICENSE)
