# TaskManager

A powerful, production-ready task scheduling library for Go, built on top of cron expressions with advanced features including concurrency control, error tracking, and flexible configuration options.

## Features

- 🕐 **Flexible Scheduling** - Standard cron expressions and fluent builder API
- 🔒 **Concurrency Control** - Configurable max concurrent tasks and overlap prevention
- 📊 **Execution Tracking** - Automatic statistics for runs, errors, and execution status
- 🎯 **Context Injection** - Support for both static and dynamic context value injection
- 🔄 **Graceful Shutdown** - Waits for running tasks to complete before shutting down
- 🚀 **Manual Triggers** - Execute tasks on-demand outside their regular schedule
- 🎛️ **Task Management** - Full CRUD operations: enable, disable, remove tasks
- 📝 **Logging** - Customizable logging output with detailed execution info
- 🌍 **Timezone Support** - Configure task execution timezone
- ⚡ **Second Precision** - Support for second-level scheduling granularity
- 🛡️ **Thread-Safe** - Safe for concurrent use across multiple goroutines
- 🔍 **Rich Metadata** - Track added time, last run, next run, and more

## Installation

```bash
go get github.com/cymoo/taskmanager
```

## Dependencies

```bash
go get github.com/robfig/cron/v3
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/cymoo/taskmanager"
)

func main() {
    // Create task manager
    tm := taskmanager.New()
    
    // Add a task that runs every 5 seconds
    tm.AddTask("hello", taskmanager.Every().Seconds(5), func(ctx context.Context) error {
        fmt.Println("Hello, World!")
        return nil
    })
    
    // Start the manager
    tm.Start()
    
    // Run for a while
    time.Sleep(30 * time.Second)
    
    // Graceful shutdown
    tm.Stop()
}
```

### Advanced Configuration

```go
// Create task manager with options
tm := taskmanager.New(
    taskmanager.WithLogger(customLogger),           // Custom logger
    taskmanager.WithLocation(location),             // Set timezone
    taskmanager.WithMaxConcurrent(5),               // Max concurrent tasks
    taskmanager.WithAllowOverlapping(false),        // Prevent overlapping
    taskmanager.WithContextValue("env", "prod"),    // Inject context values
)
```

## Schedule Expressions

### Using Builder API (Recommended)

```go
// Every second
taskmanager.Every().Second()

// Every minute
taskmanager.Every().Minute()

// Every hour
taskmanager.Every().Hour()

// Every day
taskmanager.Every().Day()

// Every N seconds
taskmanager.Every().Seconds(30)

// Every N minutes
taskmanager.Every().Minutes(15)

// Every N hours
taskmanager.Every().Hours(6)

// Every N days
taskmanager.Every().Days(2)

// Daily at specific time
taskmanager.Every().Day().At(14, 30)  // 2:30 PM daily

// Specific weekday
taskmanager.Every().Day().At(9, 0).OnWeekday(time.Monday)  // Monday 9:00 AM

// Specific day of month
taskmanager.Every().Day().At(0, 0).OnDay(1)  // 1st of every month at midnight
```

### Using Raw Cron Expressions

```go
// Format: second minute hour day month weekday
taskmanager.Cron("0 30 * * * *")     // Every hour at 30 minutes
taskmanager.Cron("0 0 2 * * *")      // Daily at 2:00 AM
taskmanager.Cron("0 */15 * * * *")   // Every 15 minutes
taskmanager.Cron("0 0 0 1 * *")      // 1st of month at midnight
taskmanager.Cron("0 0 9 * * 1")      // Every Monday at 9:00 AM
```

## Configuration Options

### WithLogger

Set a custom logger:

```go
logger := log.New(os.Stdout, "[TASK] ", log.LstdFlags)
tm := taskmanager.New(taskmanager.WithLogger(logger))
```

### WithLocation

Set timezone for task execution:

```go
location, _ := time.LoadLocation("America/New_York")
tm := taskmanager.New(taskmanager.WithLocation(location))
```

### WithMaxConcurrent

Limit maximum concurrent tasks (0 = unlimited):

```go
tm := taskmanager.New(taskmanager.WithMaxConcurrent(3))
```

### WithAllowOverlapping

Control whether the same task can run concurrently:

```go
// Prevent same task from running multiple instances
tm := taskmanager.New(taskmanager.WithAllowOverlapping(false))

// Allow same task to run concurrently
tm := taskmanager.New(taskmanager.WithAllowOverlapping(true))
```

### WithContextValue

Inject static context values available to all tasks:

```go
tm := taskmanager.New(
    taskmanager.WithContextValue("database", dbConnection),
    taskmanager.WithContextValue("cache", redisClient),
    taskmanager.WithContextValue("env", "production"),
)
```

### WithContextInjector

Dynamically inject context values per execution:

```go
tm := taskmanager.New(
    taskmanager.WithContextInjector(func(ctx context.Context, taskName string) context.Context {
        // Generate unique ID for each execution
        ctx = context.WithValue(ctx, "request_id", uuid.New().String())
        ctx = context.WithValue(ctx, "timestamp", time.Now())
        return ctx
    }),
)
```

## Task Management

### Adding Tasks

```go
err := tm.AddTask("backup", taskmanager.Every().Day().At(2, 0), func(ctx context.Context) error {
    // Perform backup logic
    return nil
})
if err != nil {
    log.Fatal(err)
}
```

### Manual Execution

Trigger a task immediately outside its schedule:

```go
err := tm.RunTaskNow("backup")
if err != nil {
    log.Printf("Manual trigger failed: %v", err)
}
```

### Disabling Tasks

Temporarily disable a task without removing it:

```go
err := tm.DisableTask("backup")
```

### Enabling Tasks

Re-enable a previously disabled task:

```go
err := tm.EnableTask("backup")
```

### Removing Tasks

Permanently remove a task:

```go
err := tm.RemoveTask("backup")
```

### Query Task Information

Get information about a specific task:

```go
taskInfo, err := tm.GetTask("backup")
if err == nil {
    fmt.Printf("Task: %s\n", taskInfo.Name)
    fmt.Printf("Schedule: %s\n", taskInfo.Schedule)
    fmt.Printf("Run Count: %d\n", taskInfo.RunCount)
    fmt.Printf("Error Count: %d\n", taskInfo.ErrorCount)
    fmt.Printf("Last Run: %s\n", taskInfo.LastRun)
    fmt.Printf("Next Run: %s\n", taskInfo.NextRun)
    fmt.Printf("Enabled: %v\n", taskInfo.Enabled)
    fmt.Printf("Running: %v\n", taskInfo.Running)
}
```

List all tasks:

```go
tasks := tm.ListTasks()
for _, task := range tasks {
    fmt.Printf("%s - Next: %s, Runs: %d, Errors: %d\n", 
        task.Name, task.NextRun, task.RunCount, task.ErrorCount)
}
```

### Statistics

Get aggregated statistics:

```go
stats := tm.GetStats()
fmt.Printf("Total Tasks: %v\n", stats["total_tasks"])
fmt.Printf("Enabled Tasks: %v\n", stats["enabled_tasks"])
fmt.Printf("Running Tasks: %v\n", stats["running_tasks"])
fmt.Printf("Total Runs: %v\n", stats["total_runs"])
fmt.Printf("Total Errors: %v\n", stats["total_errors"])
fmt.Printf("Max Concurrent: %v\n", stats["max_concurrent"])
fmt.Printf("Allow Overlapping: %v\n", stats["allow_overlapping"])
```

## Working with Context

### Retrieving Task Name

```go
tm.AddTask("example", taskmanager.Every().Minute(), func(ctx context.Context) error {
    taskName := taskmanager.GetTaskName(ctx)
    fmt.Printf("Current task: %s\n", taskName)
    return nil
})
```

### Accessing Injected Values

```go
tm.AddTask("example", taskmanager.Every().Minute(), func(ctx context.Context) error {
    db := ctx.Value("database").(*sql.DB)
    requestID := ctx.Value("request_id").(string)
    env := ctx.Value("env").(string)
    
    // Use injected values
    log.Printf("[%s] Processing in %s environment", requestID, env)
    rows, err := db.Query("SELECT * FROM users")
    // ...
    return nil
})
```

### Handling Context Cancellation

Always check for context cancellation in long-running tasks:

```go
tm.AddTask("long-running", taskmanager.Every().Hour(), func(ctx context.Context) error {
    for i := 0; i < 100; i++ {
        select {
        case <-ctx.Done():
            // Task manager is shutting down
            log.Println("Task cancelled, cleaning up...")
            return ctx.Err()
        default:
            // Continue work
            time.Sleep(time.Second)
            // Process item i
        }
    }
    return nil
})
```

### Dynamic Context Updates

Update context values at runtime:

```go
// Set or update a context value
tm.SetContextValue("feature_flag", true)

// Retrieve a context value
value := tm.GetContextValue("feature_flag")
if enabled, ok := value.(bool); ok && enabled {
    // Feature is enabled
}
```

## Error Handling

Errors returned from task functions are automatically logged and tracked:

```go
tm.AddTask("api-call", taskmanager.Every().Minutes(5), func(ctx context.Context) error {
    resp, err := http.Get("https://api.example.com/data")
    if err != nil {
        return fmt.Errorf("API call failed: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != 200 {
        return fmt.Errorf("API returned error status: %d", resp.StatusCode)
    }
    
    // Process response
    return nil
})

// Later, check for errors
taskInfo, _ := tm.GetTask("api-call")
if taskInfo.LastError != "" {
    log.Printf("Task last failed with: %s", taskInfo.LastError)
    log.Printf("Error rate: %d/%d (%.1f%%)", 
        taskInfo.ErrorCount, 
        taskInfo.RunCount,
        float64(taskInfo.ErrorCount)/float64(taskInfo.RunCount)*100)
}
```

## Graceful Shutdown

The task manager supports graceful shutdown with proper cleanup:

```go
func main() {
    tm := taskmanager.New()
    // ... add tasks
    tm.Start()
    
    // Listen for system signals
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    <-sigChan
    
    // Gracefully shutdown (waits up to 30 seconds for running tasks)
    tm.Stop()
}
```

The `Stop()` method:
1. Cancels the manager context (stops new executions)
2. Stops the cron scheduler
3. Waits for all running tasks to complete (up to 30 seconds)
4. Logs completion status

## Complete Example

See `example.go` for a comprehensive, runnable example that demonstrates:

- Multiple scheduling strategies
- Concurrency control and overlap prevention
- Error handling with simulated failures
- Context injection (static and dynamic)
- Task management operations (enable/disable)
- Statistics and monitoring
- Long-running tasks with cancellation
- Graceful shutdown handling

Run the example:

```bash
go run example.go
```

## Best Practices

### 1. Set Appropriate Concurrency Limits

```go
// For CPU-intensive tasks
tm := taskmanager.New(taskmanager.WithMaxConcurrent(runtime.NumCPU()))

// For I/O-bound tasks
tm := taskmanager.New(taskmanager.WithMaxConcurrent(20))
```

### 2. Handle Context Cancellation

Always respect context cancellation in long-running tasks:

```go
func(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Do work
        }
    }
}
```

### 3. Return Meaningful Errors

```go
func(ctx context.Context) error {
    if err := doWork(); err != nil {
        return fmt.Errorf("failed to process batch %d: %w", batchID, err)
    }
    return nil
}
```

### 4. Prevent Overlapping for Critical Tasks

```go
tm := taskmanager.New(taskmanager.WithAllowOverlapping(false))
```

### 5. Monitor Task Health

```go
// Periodically check task statistics
ticker := time.NewTicker(5 * time.Minute)
go func() {
    for range ticker.C {
        stats := tm.GetStats()
        errorRate := float64(stats["total_errors"].(int64)) / float64(stats["total_runs"].(int64))
        if errorRate > 0.1 { // More than 10% errors
            alert("High task error rate detected")
        }
    }
}()
```

### 6. Use Timeouts for External Calls

```go
tm.AddTask("api-call", schedule, func(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := client.Do(req)
    // ...
})
```

### 7. Clean Up Resources

```go
tm.AddTask("db-task", schedule, func(ctx context.Context) error {
    conn := pool.Get()
    defer conn.Close()
    
    // Use connection
    return nil
})
```

## Performance Considerations

- **Memory Usage**: Each task stores minimal metadata (~200 bytes)
- **Goroutines**: One goroutine per concurrent task execution
- **Lock Contention**: Read-write locks minimize contention on task metadata
- **Cron Performance**: Uses the highly optimized `robfig/cron` library

## Thread Safety

All TaskManager methods are thread-safe and can be called concurrently:

```go
// Safe to call from multiple goroutines
go tm.AddTask(name1, schedule1, task1)
go tm.AddTask(name2, schedule2, task2)
go tm.RunTaskNow(name1)
go tm.GetStats()
```

## Limitations

- Maximum timeout for graceful shutdown: 30 seconds
- Task names must be unique
- Cron expressions use 6 fields (seconds supported)
- Context values are copied, not referenced (use pointers for shared state)

## FAQ

**Q: Can I update a task's schedule without removing it?**  
A: Currently, you need to remove and re-add the task. A future version may support schedule updates.

**Q: What happens if a task is already running when triggered manually?**  
A: If `AllowOverlapping` is false, you'll get an error. If true, both instances will run.

**Q: How do I handle tasks that might run longer than their interval?**  
A: Set `WithAllowOverlapping(false)` to skip executions if the previous one is still running.

**Q: Can I pause the entire task manager?**  
A: Not directly. You can disable all tasks individually or stop and restart the manager.

**Q: Is it safe to modify context values during execution?**  
A: Use `SetContextValue()` to update values. Changes apply to new executions, not running ones.

## Testing

To test your tasks:

```go
func TestMyTask(t *testing.T) {
    tm := taskmanager.New()
    
    executed := false
    tm.AddTask("test", taskmanager.Every().Second(), func(ctx context.Context) error {
        executed = true
        return nil
    })
    
    tm.Start()
    time.Sleep(2 * time.Second)
    tm.Stop()
    
    if !executed {
        t.Error("Task was not executed")
    }
}
```

## License

MIT License - see [LICENSE](LICENSE) file for details.
