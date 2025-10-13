### 创建任务管理器

```go
// 不使用依赖注入
tm := taskmanager.New[any]()

// 使用依赖注入（推荐）
type App struct {
    DB    *sqlx.DB
    Redis *redis.Client
}

app := &App{...}
tm := taskmanager.New(
    taskmanager.WithDependencies(app),
    taskmanager.WithLogger[*App](customLogger),
    taskmanager.WithLocation[*App](time.UTC),
    taskmanager.WithMaxConcurrent## 注意事项

- 任务名称必须唯一
- Cron 表达式格式错误会在 AddTask 时返回错误
- Stop() 会等待最多 30 秒让任务完成
- 禁用的任务不会被调度执行，但仍保留在系统中
- 默认情况下，同一任务不会重叠执行（上次未完成时会跳过本次调度）
- 设置 `MaxConcurrent` 会限制所有任务的总并发数
- 正在运行的任务会在 `TaskInfo.Running` 字段中标记
- 错误信息会保存在 `TaskInfo.LastError` 中，成功执行后会清空

## 性能考虑

### 内存使用
- 每个任务约占用 ~1KB 内存（不含任务函数的闭包）
- 建议单个实例管理不超过 10000 个任务

### CPU 使用
- Cron 调度器使用极少 CPU
- 主要开销来自任务本身的执行
- 建议使用 `MaxConcurrent` 限制并发，避免 CPU 过载

### 锁竞争
- 读操作（ListTasks, GetTask）使用读锁，可并发
- 写操作（AddTask, RemoveTask）使用写锁，会阻塞
- 任务执行时只在开始和结束时短暂持锁，不影响并发# Go 任务管理系统

一个简单优雅的 Go 任务调度系统，基于 cron 表达式，可与 net/http 无缝集成。

## 特性

- ✅ 支持标准 cron 表达式（含秒级精度）
- ✅ **泛型依赖注入**（类型安全、优雅简洁）
- ✅ **并行执行多个不同任务**
- ✅ **可配置最大并发数限制**
- ✅ **防止同一任务重叠执行（可配置）**
- ✅ 任务的启用/禁用控制
- ✅ 立即运行任务
- ✅ 任务统计（运行次数、错误次数、执行时间、上次运行时间、下次运行时间）
- ✅ 优雅关闭（等待正在执行的任务完成）
- ✅ 线程安全
- ✅ HTTP API 管理接口
- ✅ Context 支持，便于取消和超时控制

## 安装

```bash
go get github.com/robfig/cron/v3
```

## 快速开始

### 基本用法（带依赖注入）

```go
package main

import (
    "context"
    "log"
    "time"
    "your-module/taskmanager"
    "github.com/jmoiron/sqlx"
    "github.com/redis/go-redis/v9"
)

// 定义应用依赖
type App struct {
    DB    *sqlx.DB
    Redis *redis.Client
}

func main() {
    // 初始化依赖
    app := &App{
        DB:    initDB(),
        Redis: initRedis(),
    }
    
    // 创建任务管理器，注入依赖
    tm := taskmanager.New(
        taskmanager.WithDependencies(app),
    )
    
    // 添加任务（推荐方式）
    tm.AddTaskWithDeps("clean-data", "0 0 3 * * *", 
        func(ctx context.Context, app *App) error {
            // 直接使用注入的依赖，类型安全
            _, err := app.DB.ExecContext(ctx, "DELETE FROM old_data")
            return err
        },
    )
    
    // 启动
    tm.Start()
    defer tm.Stop()
    
    // 保持运行
    select {}
}
```

### Cron 表达式格式

系统支持 6 位 cron 表达式（含秒）：

```
秒 分 时 日 月 周

字段         允许值                  特殊字符
秒           0-59                   * / , -
分           0-59                   * / , -
时           0-23                   * / , -
日           1-31                   * / , - ?
月           1-12 或 JAN-DEC        * / , -
周           0-6 或 SUN-SAT         * / , - ?
```

### 常用表达式示例

```go
// 每天凌晨 3 点
"0 0 3 * * *"

// 每 5 分钟
"0 */5 * * * *"

// 每小时整点
"0 0 * * * *"

// 每周一早上 8 点
"0 0 8 * * MON"

// 每 30 秒
"*/30 * * * * *"

// 工作日上午 9 点
"0 0 9 * * MON-FRI"

// 每月 1 号凌晨 2 点
"0 0 2 1 * *"
```

## API 参考

### 创建任务管理器

```go
// 不使用依赖注入
tm := taskmanager.New[any]()

// 使用依赖注入（推荐）
type App struct {
    DB    *sqlx.DB
    Redis *redis.Client
}

app := &App{...}
tm := taskmanager.New(
    taskmanager.WithDependencies(app),
    taskmanager.WithLogger[*App](customLogger),
    taskmanager.WithLocation[*App](time.UTC),
    taskmanager.WithMaxConcurrent[*App](10),        // 最多同时运行 10 个任务
    taskmanager.WithAllowOverlapping[*App](false),  // 禁止同一任务重叠执行
)
```

### 任务管理

```go
// 方式 1: 使用依赖注入（推荐）
tm.AddTaskWithDeps("task-name", "0 0 * * * *", 
    func(ctx context.Context, app *App) error {
        // 直接使用依赖，类型安全
        return app.DB.Ping()
    },
)

// 方式 2: 不使用依赖（兼容）
tm.AddTask("simple-task", "0 0 * * * *", 
    func(ctx context.Context) error {
        // 简单任务逻辑
        return nil
    },
)

// 移除任务
err := tm.RemoveTask("task-name")

// 启用/禁用任务
err := tm.EnableTask("task-name")
err := tm.DisableTask("task-name")

// 立即运行任务
err := tm.RunTaskNow("task-name")

// 获取任务信息
info, err := tm.GetTask("task-name")

// 列出所有任务
tasks := tm.ListTasks()

// 获取统计信息
stats := tm.GetStats()
```

### 生命周期

```go
// 启动任务管理器
tm.Start()

// 停止任务管理器（优雅关闭）
tm.Stop()

// 检查是否运行中
isRunning := tm.IsRunning()
```

## HTTP API 集成

### API 端点

```
GET    /api/tasks              - 获取所有任务列表
GET    /api/tasks/{name}       - 获取指定任务信息
DELETE /api/tasks/{name}       - 删除指定任务
POST   /api/tasks/run/{name}   - 立即运行指定任务
POST   /api/tasks/enable/{name}  - 启用指定任务
POST   /api/tasks/disable/{name} - 禁用指定任务
GET    /api/stats              - 获取任务管理器统计信息
GET    /health                 - 健康检查
```

### 使用示例

```bash
# 获取所有任务
curl http://localhost:8080/api/tasks

# 获取单个任务
curl http://localhost:8080/api/tasks/clean-data

# 立即运行任务
curl -X POST http://localhost:8080/api/tasks/run/clean-data

# 禁用任务
curl -X POST http://localhost:8080/api/tasks/disable/clean-data

# 启用任务
curl -X POST http://localhost:8080/api/tasks/enable/clean-data

# 删除任务
curl -X DELETE http://localhost:8080/api/tasks/clean-data

# 健康检查
curl http://localhost:8080/health

# 获取统计信息
curl http://localhost:8080/api/stats
```

## 并发执行说明

### 默认行为（完全并行）

```go
// 不设置限制，所有任务都会并行执行
tm := taskmanager.New()
```

在这种模式下：
- ✅ 不同的任务会并行执行
- ✅ 同一个任务默认**不会**重叠执行（上次未完成时跳过本次）
- ✅ 无并发数限制

### 限制最大并发数

```go
// 最多同时运行 5 个任务
tm := taskmanager.New(
    taskmanager.WithMaxConcurrent(5),
)
```

使用场景：
- 保护系统资源（CPU、内存、数据库连接等）
- 防止任务过多导致系统负载过高
- 控制外部 API 调用频率

### 允许任务重叠执行

```go
// 允许同一任务在上次未完成时再次启动
tm := taskmanager.New(
    taskmanager.WithAllowOverlapping(true),
)
```

⚠️ **注意**：通常不建议开启，除非你的任务是幂等的且确实需要并发执行。

### 最佳实践配置

```go
tm := taskmanager.New(
    taskmanager.WithMaxConcurrent(10),       // 限制总并发
    taskmanager.WithAllowOverlapping(false), // 防止重叠（默认）
)
```

## 并发示例

```go
// 这些任务会并行执行
tm.AddTask("task1", "*/10 * * * * *", longRunningTask1)
tm.AddTask("task2", "*/10 * * * * *", longRunningTask2)
tm.AddTask("task3", "*/10 * * * * *", longRunningTask3)

// 如果设置了 WithMaxConcurrent(2)，则最多同时运行 2 个任务
// 第 3 个任务会等待前面的任务完成后再执行
```

## 高级用法

### 依赖注入（推荐）

```go
// 定义应用依赖
type App struct {
    Config *Config
    DB     *sqlx.DB
    Redis  *redis.Client
    Logger *log.Logger
}

// 创建任务管理器并注入依赖
tm := taskmanager.New(
    taskmanager.WithDependencies(&App{...}),
)

// 任务中直接使用依赖，类型安全
tm.AddTaskWithDeps("clean-db", "0 0 2 * * *", 
    func(ctx context.Context, app *App) error {
        app.Logger.Println("Cleaning database...")
        
        // 使用数据库
        result, err := app.DB.ExecContext(ctx, 
            "DELETE FROM logs WHERE created_at < NOW() - INTERVAL '30 days'")
        if err != nil {
            return err
        }
        
        // 使用 Redis
        app.Redis.Del(ctx, "temp:cache")
        
        return nil
    },
)
```

**为什么不用 Context.Value？**
- ✅ 类型安全，编译时检查
- ✅ 代码简洁，无需类型断言
- ✅ IDE 自动补全支持
- ✅ 零运行时开销
- ✅ 更符合 Go 最佳实践

### Context 支持

```go
tm.AddTaskWithDeps("timeout-task", "0 * * * * *", 
    func(ctx context.Context, app *App) error {
        // 使用 context 进行超时控制
        ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
        defer cancel()
        
        return app.DB.PingContext(ctx)
    },
)
```

### 错误处理

```go
tm.AddTaskWithDeps("error-task", "0 * * * * *", 
    func(ctx context.Context, app *App) error {
        if err := doSomething(app.DB); err != nil {
            // 错误会被自动记录到 TaskInfo.ErrorCount 和 TaskInfo.LastError
            return fmt.Errorf("failed to do something: %w", err)
        }
        return nil
    },
)
```

### 自定义日志

```go
logger := log.New(os.Stdout, "[MyApp] ", log.LstdFlags|log.Lshortfile)
tm := taskmanager.New(
    taskmanager.WithLogger[*App](logger),
    taskmanager.WithDependencies(app),
)
```

### 时区设置

```go
loc, _ := time.LoadLocation("Asia/Shanghai")
tm := taskmanager.New(
    taskmanager.WithLocation[*App](loc),
    taskmanager.WithDependencies(app),
)
```

## 最佳实践

1. **优雅关闭**: 始终在应用退出时调用 `tm.Stop()` 以确保正在运行的任务完成

2. **错误处理**: 任务函数应返回错误，系统会自动记录错误次数和最后错误信息

3. **Context 使用**: 在长时间运行的任务中检查 context 以支持取消操作

4. **幂等性**: 设计任务时应考虑幂等性，因为任务可能因系统重启而重复执行

5. **资源管理**: 在任务中使用 defer 确保资源正确释放

6. **并发控制**: 
   - 根据系统资源合理设置 `MaxConcurrent`
   - 默认禁止同一任务重叠，避免资源竞争
   - 对于可能长时间运行的任务，建议设置超时

7. **监控**: 定期检查任务统计信息（`GetStats()`），监控错误率和执行情况

## 注意事项

- 任务名称必须唯一
- Cron 表达式格式错误会在 AddTask 时返回错误
- Stop() 会等待最多 30 秒让任务完成
- 禁用的任务不会被调度执行，但仍保留在系统中

## License

MIT