package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cymoo/pebble/pkg/taskmanager"
)

func main() {
	// Create custom logger
	logger := log.New(os.Stdout, "[TaskManager] ", log.LstdFlags)

	// Load timezone
	location, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Fatalf("Failed to load timezone: %v", err)
	}

	// Create task manager with various options
	tm := taskmanager.New(
		taskmanager.WithLogger(logger),
		taskmanager.WithLocation(location),
		taskmanager.WithMaxConcurrent(3),        // Max 3 concurrent tasks
		taskmanager.WithAllowOverlapping(false), // Prevent overlapping executions
		taskmanager.WithContextValue("app", "demo"),
		taskmanager.WithContextInjector(func(ctx context.Context, taskName string) context.Context {
			// Dynamically inject request ID
			return context.WithValue(ctx, "request_id", fmt.Sprintf("%s-%d", taskName, time.Now().Unix()))
		}),
	)

	// Example 1: Data cleanup task - runs every 5 seconds
	err = tm.AddTask("cleanup", taskmanager.Every().Seconds(5), func(ctx context.Context) error {
		taskName := taskmanager.GetTaskName(ctx)
		app := ctx.Value("app")
		requestID := ctx.Value("request_id")

		fmt.Printf("[%s] Starting data cleanup... (app=%v, request_id=%v)\n", taskName, app, requestID)
		time.Sleep(2 * time.Second) // Simulate work
		fmt.Printf("[%s] Cleanup completed\n", taskName)
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to add cleanup task: %v", err)
	}

	// Example 2: Data sync task - runs every 10 seconds (may fail)
	err = tm.AddTask("sync", taskmanager.Every().Seconds(10), func(ctx context.Context) error {
		taskName := taskmanager.GetTaskName(ctx)
		fmt.Printf("[%s] Starting data synchronization...\n", taskName)

		// Simulate 30% failure rate
		if time.Now().Unix()%3 == 0 {
			return fmt.Errorf("network connection timeout")
		}

		time.Sleep(1 * time.Second)
		fmt.Printf("[%s] Sync completed successfully\n", taskName)
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to add sync task: %v", err)
	}

	// Example 3: Report generation - runs every minute
	err = tm.AddTask("report", taskmanager.Every().Minute(), func(ctx context.Context) error {
		taskName := taskmanager.GetTaskName(ctx)
		fmt.Printf("[%s] Generating minute report...\n", taskName)

		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(3 * time.Second):
			fmt.Printf("[%s] Report generation completed\n", taskName)
			return nil
		}
	})
	if err != nil {
		log.Fatalf("Failed to add report task: %v", err)
	}

	// Example 4: Daily backup at 2:00 AM
	err = tm.AddTask("backup", taskmanager.Every().Day().At(2, 0), func(ctx context.Context) error {
		taskName := taskmanager.GetTaskName(ctx)
		fmt.Printf("[%s] Starting daily backup...\n", taskName)
		time.Sleep(1 * time.Second)
		fmt.Printf("[%s] Backup completed\n", taskName)
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to add backup task: %v", err)
	}

	// Example 5: Health check using raw cron expression - every 15 minutes
	err = tm.AddTask("health-check", taskmanager.Cron("0 */15 * * * *"), func(ctx context.Context) error {
		taskName := taskmanager.GetTaskName(ctx)
		fmt.Printf("[%s] Performing health check...\n", taskName)
		time.Sleep(500 * time.Millisecond)
		fmt.Printf("[%s] System healthy\n", taskName)
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to add health check task: %v", err)
	}

	// Example 6: Long-running task with context cancellation handling
	err = tm.AddTask("long-task", taskmanager.Every().Minutes(2), func(ctx context.Context) error {
		taskName := taskmanager.GetTaskName(ctx)
		fmt.Printf("[%s] Starting long-running task...\n", taskName)

		// Simulate long processing with proper cancellation handling
		for i := 0; i < 10; i++ {
			select {
			case <-ctx.Done():
				fmt.Printf("[%s] Task cancelled, cleaning up...\n", taskName)
				return ctx.Err()
			default:
				fmt.Printf("[%s] Processing step %d/10...\n", taskName, i+1)
				time.Sleep(500 * time.Millisecond)
			}
		}

		fmt.Printf("[%s] Long task completed\n", taskName)
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to add long-running task: %v", err)
	}

	// Start the task manager
	tm.Start()
	fmt.Println("\n✓ Task manager started. Press Ctrl+C to stop...")
	fmt.Println(strings.Repeat("=", 60))

	// Manually trigger sync task after 3 seconds
	go func() {
		time.Sleep(3 * time.Second)
		fmt.Println("\n>>> Manually triggering sync task <<<")
		if err := tm.RunTaskNow("sync"); err != nil {
			log.Printf("Manual trigger failed: %v", err)
		}
	}()

	// Demonstrate task management operations
	go func() {
		time.Sleep(15 * time.Second)

		// Disable cleanup task
		fmt.Println("\n>>> Disabling cleanup task <<<")
		if err := tm.DisableTask("cleanup"); err != nil {
			log.Printf("Failed to disable task: %v", err)
		}

		time.Sleep(10 * time.Second)

		// Re-enable cleanup task
		fmt.Println("\n>>> Re-enabling cleanup task <<<")
		if err := tm.EnableTask("cleanup"); err != nil {
			log.Printf("Failed to enable task: %v", err)
		}
	}()

	// Demonstrate dynamic context value updates
	go func() {
		ticker := time.NewTicker(25 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			tm.SetContextValue("last_update", time.Now().Format(time.RFC3339))
			fmt.Println("\n>>> Updated context value: last_update <<<")
		}
	}()

	// Display statistics periodically
	go func() {
		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			displayStats(tm)
		}
	}()

	// Listen for system signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n\nReceived shutdown signal, stopping task manager...")
	tm.Stop()
	fmt.Println("✓ Task manager stopped gracefully")
}

// displayStats shows current task manager statistics
func displayStats(tm *taskmanager.TaskManager) {
	fmt.Println("\n" + string(make([]byte, 70)))
	fmt.Println("TASK MANAGER STATISTICS")
	fmt.Println(string(make([]byte, 70)))

	stats := tm.GetStats()
	fmt.Printf("Total Tasks:       %v\n", stats["total_tasks"])
	fmt.Printf("Enabled Tasks:     %v\n", stats["enabled_tasks"])
	fmt.Printf("Running Tasks:     %v\n", stats["running_tasks"])
	fmt.Printf("Total Executions:  %v\n", stats["total_runs"])
	fmt.Printf("Total Errors:      %v\n", stats["total_errors"])
	fmt.Printf("Max Concurrent:    %v\n", stats["max_concurrent"])
	fmt.Printf("Allow Overlapping: %v\n", stats["allow_overlapping"])

	fmt.Println("\nPER-TASK DETAILS")
	fmt.Println(string(make([]byte, 70)))

	tasks := tm.ListTasks()
	for _, task := range tasks {
		fmt.Printf("\n📌 Task: %s\n", task.Name)
		fmt.Printf("   Schedule:     %s\n", task.Schedule)
		fmt.Printf("   Status:       Enabled=%v, Running=%v\n", task.Enabled, task.Running)
		fmt.Printf("   Executions:   %d (Errors: %d)\n", task.RunCount, task.ErrorCount)
		fmt.Printf("   Added At:     %s\n", task.AddedAt.Format("2006-01-02 15:04:05"))

		if !task.LastRun.IsZero() {
			fmt.Printf("   Last Run:     %s\n", task.LastRun.Format("2006-01-02 15:04:05"))
		}

		if !task.NextRun.IsZero() {
			fmt.Printf("   Next Run:     %s\n", task.NextRun.Format("2006-01-02 15:04:05"))
		}

		if task.LastError != "" {
			fmt.Printf("   ⚠️  Last Error: %s\n", task.LastError)
		}

		// Calculate success rate
		if task.RunCount > 0 {
			successRate := float64(task.RunCount-task.ErrorCount) / float64(task.RunCount) * 100
			fmt.Printf("   Success Rate: %.1f%%\n", successRate)
		}
	}

	fmt.Println("\n" + string(make([]byte, 70)) + "\n")
}
