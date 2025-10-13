package taskmanager

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestNew verifies TaskManager creation with various options
func TestNew(t *testing.T) {
	tests := []struct {
		name string
		opts []Option
		want func(*TaskManager) bool
	}{
		{
			name: "default configuration",
			opts: nil,
			want: func(tm *TaskManager) bool {
				return tm != nil && tm.IsRunning() && tm.maxConcurrent == 0 && !tm.allowOverlapping
			},
		},
		{
			name: "with max concurrent",
			opts: []Option{WithMaxConcurrent(5)},
			want: func(tm *TaskManager) bool {
				return tm.maxConcurrent == 5 && tm.semaphore != nil
			},
		},
		{
			name: "with negative max concurrent",
			opts: []Option{WithMaxConcurrent(-1)},
			want: func(tm *TaskManager) bool {
				return tm.maxConcurrent == 0 && tm.semaphore == nil
			},
		},
		{
			name: "with allow overlapping",
			opts: []Option{WithAllowOverlapping(true)},
			want: func(tm *TaskManager) bool {
				return tm.allowOverlapping
			},
		},
		{
			name: "with context value",
			opts: []Option{WithContextValue("key1", "value1")},
			want: func(tm *TaskManager) bool {
				return tm.GetContextValue("key1") == "value1"
			},
		},
		{
			name: "with empty context key",
			opts: []Option{WithContextValue("", "value1")},
			want: func(tm *TaskManager) bool {
				return tm.GetContextValue("") == nil
			},
		},
		{
			name: "with custom logger",
			opts: []Option{WithLogger(log.New(os.Stdout, "[TEST] ", log.LstdFlags))},
			want: func(tm *TaskManager) bool {
				return tm.logger != nil
			},
		},
		{
			name: "with nil logger",
			opts: []Option{WithLogger(nil)},
			want: func(tm *TaskManager) bool {
				return tm.logger == log.Default()
			},
		},
		{
			name: "with location",
			opts: []Option{WithLocation(time.UTC)},
			want: func(tm *TaskManager) bool {
				return tm.cron != nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := New(tt.opts...)
			defer tm.Stop()

			if !tt.want(tm) {
				t.Errorf("TaskManager validation failed for test: %s", tt.name)
			}
		})
	}
}

// TestAddTask verifies task addition with various scenarios
func TestAddTask(t *testing.T) {
	tests := []struct {
		name      string
		taskName  string
		schedule  Schedule
		task      Task
		wantError bool
	}{
		{
			name:      "valid task",
			taskName:  "test-task",
			schedule:  Every().Minute(),
			task:      func(ctx context.Context) error { return nil },
			wantError: false,
		},
		{
			name:      "empty task name",
			taskName:  "",
			schedule:  Every().Minute(),
			task:      func(ctx context.Context) error { return nil },
			wantError: true,
		},
		{
			name:      "nil schedule",
			taskName:  "test-task",
			schedule:  nil,
			task:      func(ctx context.Context) error { return nil },
			wantError: true,
		},
		{
			name:      "nil task function",
			taskName:  "test-task",
			schedule:  Every().Minute(),
			task:      nil,
			wantError: true,
		},
		{
			name:      "invalid cron expression",
			taskName:  "test-task",
			schedule:  Cron("invalid cron"),
			task:      func(ctx context.Context) error { return nil },
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := New()
			defer tm.Stop()

			err := tm.AddTask(tt.taskName, tt.schedule, tt.task)
			if (err != nil) != tt.wantError {
				t.Errorf("AddTask() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError {
				info, err := tm.GetTask(tt.taskName)
				if err != nil {
					t.Errorf("GetTask() error = %v", err)
				}
				if info.Name != tt.taskName {
					t.Errorf("Task name = %v, want %v", info.Name, tt.taskName)
				}
			}
		})
	}
}

// TestAddDuplicateTask verifies duplicate task prevention
func TestAddDuplicateTask(t *testing.T) {
	tm := New()
	defer tm.Stop()

	task := func(ctx context.Context) error { return nil }
	schedule := Every().Minute()

	err := tm.AddTask("duplicate", schedule, task)
	if err != nil {
		t.Fatalf("First AddTask() failed: %v", err)
	}

	err = tm.AddTask("duplicate", schedule, task)
	if err == nil {
		t.Error("Expected error when adding duplicate task, got nil")
	}
}

// TestTaskExecution verifies tasks execute correctly
func TestTaskExecution(t *testing.T) {
	tm := New()
	tm.Start()
	defer tm.Stop()

	var counter atomic.Int32
	task := func(ctx context.Context) error {
		counter.Add(1)
		return nil
	}

	// Task should run every second
	err := tm.AddTask("counter", Every().Second(), task)
	if err != nil {
		t.Fatalf("AddTask() failed: %v", err)
	}

	// Wait for at least 2 executions
	time.Sleep(2500 * time.Millisecond)

	count := counter.Load()
	if count < 2 {
		t.Errorf("Task executed %d times, expected at least 2", count)
	}
}

// TestTaskExecutionWithError verifies error handling
func TestTaskExecutionWithError(t *testing.T) {
	tm := New()
	tm.Start()
	defer tm.Stop()

	expectedErr := errors.New("task failed")
	task := func(ctx context.Context) error {
		return expectedErr
	}

	err := tm.AddTask("error-task", Every().Second(), task)
	if err != nil {
		t.Fatalf("AddTask() failed: %v", err)
	}

	// Wait for execution
	time.Sleep(1500 * time.Millisecond)

	info, err := tm.GetTask("error-task")
	if err != nil {
		t.Fatalf("GetTask() failed: %v", err)
	}

	if info.ErrorCount == 0 {
		t.Error("Expected ErrorCount > 0, got 0")
	}

	if info.LastError != expectedErr.Error() {
		t.Errorf("LastError = %v, want %v", info.LastError, expectedErr.Error())
	}
}

// TestRunTaskNow verifies manual task execution
func TestRunTaskNow(t *testing.T) {
	tm := New()
	tm.Start()
	defer tm.Stop()

	var executed atomic.Bool
	task := func(ctx context.Context) error {
		executed.Store(true)
		return nil
	}

	// Add task with a schedule that won't trigger during test
	err := tm.AddTask("manual", Cron("0 0 0 1 1 *"), task)
	if err != nil {
		t.Fatalf("AddTask() failed: %v", err)
	}

	err = tm.RunTaskNow("manual")
	if err != nil {
		t.Fatalf("RunTaskNow() failed: %v", err)
	}

	// Wait for execution
	time.Sleep(100 * time.Millisecond)

	if !executed.Load() {
		t.Error("Task was not executed")
	}
}

// TestRunTaskNowErrors verifies RunTaskNow error cases
func TestRunTaskNowErrors(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*TaskManager)
		taskName  string
		wantError bool
	}{
		{
			name:      "empty task name",
			setup:     func(tm *TaskManager) {},
			taskName:  "",
			wantError: true,
		},
		{
			name:      "non-existent task",
			setup:     func(tm *TaskManager) {},
			taskName:  "non-existent",
			wantError: true,
		},
		{
			name: "disabled task",
			setup: func(tm *TaskManager) {
				tm.AddTask("disabled", Every().Minute(), func(ctx context.Context) error { return nil })
				tm.DisableTask("disabled")
			},
			taskName:  "disabled",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := New()
			defer tm.Stop()
			tt.setup(tm)

			err := tm.RunTaskNow(tt.taskName)
			if (err != nil) != tt.wantError {
				t.Errorf("RunTaskNow() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestDisableEnableTask verifies task enable/disable functionality
func TestDisableEnableTask(t *testing.T) {
	tm := New()
	tm.Start()
	defer tm.Stop()

	var counter atomic.Int32
	task := func(ctx context.Context) error {
		counter.Add(1)
		return nil
	}

	err := tm.AddTask("toggle", Every().Second(), task)
	if err != nil {
		t.Fatalf("AddTask() failed: %v", err)
	}

	// Let it run once
	time.Sleep(1200 * time.Millisecond)
	firstCount := counter.Load()

	// Disable task
	err = tm.DisableTask("toggle")
	if err != nil {
		t.Fatalf("DisableTask() failed: %v", err)
	}

	// Wait and verify it doesn't run
	time.Sleep(1500 * time.Millisecond)
	if counter.Load() != firstCount {
		t.Error("Task executed while disabled")
	}

	// Re-enable task
	err = tm.EnableTask("toggle")
	if err != nil {
		t.Fatalf("EnableTask() failed: %v", err)
	}

	// Verify it runs again
	time.Sleep(1200 * time.Millisecond)
	if counter.Load() == firstCount {
		t.Error("Task did not execute after being re-enabled")
	}
}

// TestRemoveTask verifies task removal
func TestRemoveTask(t *testing.T) {
	tm := New()
	defer tm.Stop()

	task := func(ctx context.Context) error { return nil }
	err := tm.AddTask("removable", Every().Minute(), task)
	if err != nil {
		t.Fatalf("AddTask() failed: %v", err)
	}

	err = tm.RemoveTask("removable")
	if err != nil {
		t.Fatalf("RemoveTask() failed: %v", err)
	}

	_, err = tm.GetTask("removable")
	if err == nil {
		t.Error("Expected error getting removed task, got nil")
	}

	// Test removing non-existent task
	err = tm.RemoveTask("non-existent")
	if err == nil {
		t.Error("Expected error removing non-existent task, got nil")
	}

	// Test removing with empty name
	err = tm.RemoveTask("")
	if err == nil {
		t.Error("Expected error removing task with empty name, got nil")
	}
}

// TestListTasks verifies task listing
func TestListTasks(t *testing.T) {
	tm := New()
	defer tm.Stop()

	task := func(ctx context.Context) error { return nil }

	names := []string{"task1", "task2", "task3"}
	for _, name := range names {
		err := tm.AddTask(name, Every().Minute(), task)
		if err != nil {
			t.Fatalf("AddTask() failed for %s: %v", name, err)
		}
	}

	tasks := tm.ListTasks()
	if len(tasks) != len(names) {
		t.Errorf("ListTasks() returned %d tasks, want %d", len(tasks), len(names))
	}

	// Verify all tasks are present
	taskMap := make(map[string]bool)
	for _, info := range tasks {
		taskMap[info.Name] = true
	}

	for _, name := range names {
		if !taskMap[name] {
			t.Errorf("Task %s not found in list", name)
		}
	}
}

// TestMaxConcurrent verifies concurrency limiting
func TestMaxConcurrent(t *testing.T) {
	maxConcurrent := 2
	tm := New(WithMaxConcurrent(maxConcurrent))
	tm.Start()
	defer tm.Stop()

	var concurrentCount atomic.Int32
	var maxObserved atomic.Int32

	task := func(ctx context.Context) error {
		current := concurrentCount.Add(1)

		// Track maximum concurrent executions
		for {
			max := maxObserved.Load()
			if current <= max || maxObserved.CompareAndSwap(max, current) {
				break
			}
		}

		time.Sleep(500 * time.Millisecond)
		concurrentCount.Add(-1)
		return nil
	}

	// Add multiple tasks that will trigger simultaneously
	for i := 0; i < 5; i++ {
		err := tm.AddTask(fmt.Sprintf("concurrent-%d", i), Every().Second(), task)
		if err != nil {
			t.Fatalf("AddTask() failed: %v", err)
		}
	}

	// Wait for executions
	time.Sleep(2 * time.Second)

	max := maxObserved.Load()
	if max > int32(maxConcurrent) {
		t.Errorf("Max concurrent tasks = %d, want <= %d", max, maxConcurrent)
	}
}

// TestOverlapPrevention verifies overlap prevention
func TestOverlapPrevention(t *testing.T) {
	tm := New(WithAllowOverlapping(false))
	tm.Start()
	defer tm.Stop()

	var executing atomic.Bool
	var overlapDetected atomic.Bool

	task := func(ctx context.Context) error {
		if !executing.CompareAndSwap(false, true) {
			overlapDetected.Store(true)
			return nil
		}
		time.Sleep(1500 * time.Millisecond)
		executing.Store(false)
		return nil
	}

	err := tm.AddTask("no-overlap", Every().Second(), task)
	if err != nil {
		t.Fatalf("AddTask() failed: %v", err)
	}

	// Wait for multiple potential executions
	time.Sleep(3 * time.Second)

	if overlapDetected.Load() {
		t.Error("Task overlap detected when overlapping is disabled")
	}
}

// TestAllowOverlapping verifies overlapping execution when enabled
func TestAllowOverlapping(t *testing.T) {
	tm := New(WithAllowOverlapping(true))
	tm.Start()
	defer tm.Stop()

	var concurrentCount atomic.Int32
	var hadConcurrent atomic.Bool

	task := func(ctx context.Context) error {
		count := concurrentCount.Add(1)
		if count > 1 {
			hadConcurrent.Store(true)
		}
		time.Sleep(1500 * time.Millisecond)
		concurrentCount.Add(-1)
		return nil
	}

	err := tm.AddTask("allow-overlap", Every().Second(), task)
	if err != nil {
		t.Fatalf("AddTask() failed: %v", err)
	}

	// Wait for multiple executions
	time.Sleep(3 * time.Second)

	if !hadConcurrent.Load() {
		t.Error("Expected concurrent execution when overlapping is allowed")
	}
}

// TestContextInjection verifies context value injection
func TestContextInjection(t *testing.T) {
	staticValue := "static-value"
	dynamicValue := "dynamic-value"

	tm := New(
		WithContextValue("static-key", staticValue),
		WithContextInjector(func(ctx context.Context, taskName string) context.Context {
			return context.WithValue(ctx, "dynamic-key", dynamicValue)
		}),
	)
	tm.Start()
	defer tm.Stop()

	var receivedStatic, receivedDynamic, receivedTaskName string
	var done atomic.Bool

	task := func(ctx context.Context) error {
		if v, ok := ctx.Value("static-key").(string); ok {
			receivedStatic = v
		}
		if v, ok := ctx.Value("dynamic-key").(string); ok {
			receivedDynamic = v
		}
		receivedTaskName = GetTaskName(ctx)
		done.Store(true)
		return nil
	}

	err := tm.AddTask("context-test", Every().Second(), task)
	if err != nil {
		t.Fatalf("AddTask() failed: %v", err)
	}

	// Wait for execution
	for i := 0; i < 20 && !done.Load(); i++ {
		time.Sleep(100 * time.Millisecond)
	}

	if receivedStatic != staticValue {
		t.Errorf("Static context value = %v, want %v", receivedStatic, staticValue)
	}

	if receivedDynamic != dynamicValue {
		t.Errorf("Dynamic context value = %v, want %v", receivedDynamic, dynamicValue)
	}

	if receivedTaskName != "context-test" {
		t.Errorf("Task name = %v, want %v", receivedTaskName, "context-test")
	}
}

// TestSetGetContextValue verifies runtime context value management
func TestSetGetContextValue(t *testing.T) {
	tm := New()
	defer tm.Stop()

	// Test setting and getting
	tm.SetContextValue("key1", "value1")
	if v := tm.GetContextValue("key1"); v != "value1" {
		t.Errorf("GetContextValue() = %v, want value1", v)
	}

	// Test updating
	tm.SetContextValue("key1", "value2")
	if v := tm.GetContextValue("key1"); v != "value2" {
		t.Errorf("GetContextValue() = %v, want value2", v)
	}

	// Test non-existent key
	if v := tm.GetContextValue("non-existent"); v != nil {
		t.Errorf("GetContextValue() = %v, want nil", v)
	}

	// Test empty key
	tm.SetContextValue("", "value")
	if v := tm.GetContextValue(""); v != nil {
		t.Errorf("GetContextValue() with empty key = %v, want nil", v)
	}
}

// TestGetStats verifies statistics collection
func TestGetStats(t *testing.T) {
	tm := New(WithMaxConcurrent(5), WithAllowOverlapping(true))
	tm.Start()
	defer tm.Stop()

	var counter atomic.Int32
	successTask := func(ctx context.Context) error {
		counter.Add(1)
		return nil
	}

	errorTask := func(ctx context.Context) error {
		return errors.New("task error")
	}

	err := tm.AddTask("success-task", Every().Second(), successTask)
	if err != nil {
		t.Fatalf("AddTask() failed: %v", err)
	}

	err = tm.AddTask("error-task", Every().Second(), errorTask)
	if err != nil {
		t.Fatalf("AddTask() failed: %v", err)
	}

	// Wait for some executions
	time.Sleep(2500 * time.Millisecond)

	stats := tm.GetStats()

	if stats["total_tasks"].(int) != 2 {
		t.Errorf("total_tasks = %v, want 2", stats["total_tasks"])
	}

	if stats["enabled_tasks"].(int) != 2 {
		t.Errorf("enabled_tasks = %v, want 2", stats["enabled_tasks"])
	}

	if stats["max_concurrent"].(int) != 5 {
		t.Errorf("max_concurrent = %v, want 5", stats["max_concurrent"])
	}

	if stats["allow_overlapping"].(bool) != true {
		t.Errorf("allow_overlapping = %v, want true", stats["allow_overlapping"])
	}

	totalRuns := stats["total_runs"].(int64)
	if totalRuns < 2 {
		t.Errorf("total_runs = %v, want >= 2", totalRuns)
	}

	totalErrors := stats["total_errors"].(int64)
	if totalErrors == 0 {
		t.Error("expected some errors, got 0")
	}
}

// TestGracefulShutdown verifies graceful shutdown
func TestGracefulShutdown(t *testing.T) {
	tm := New()
	tm.Start()

	var started atomic.Bool
	var completed atomic.Bool

	task := func(ctx context.Context) error {
		started.Store(true)
		time.Sleep(500 * time.Millisecond)
		completed.Store(true)
		return nil
	}

	err := tm.AddTask("long-task", Every().Second(), task)
	if err != nil {
		t.Fatalf("AddTask() failed: %v", err)
	}

	// Wait for task to start
	time.Sleep(1200 * time.Millisecond)

	if !started.Load() {
		t.Fatal("Task did not start")
	}

	// Stop should wait for task to complete
	tm.Stop()

	if !completed.Load() {
		t.Error("Task was not allowed to complete during shutdown")
	}

	if tm.IsRunning() {
		t.Error("TaskManager is still running after Stop()")
	}
}

// TestScheduleBuilder verifies schedule building
func TestScheduleBuilder(t *testing.T) {
	tests := []struct {
		name     string
		schedule Schedule
		want     string
	}{
		{
			name:     "every second",
			schedule: Every().Second(),
			want:     "* * * * * *",
		},
		{
			name:     "every minute",
			schedule: Every().Minute(),
			want:     "0 * * * * *",
		},
		{
			name:     "every hour",
			schedule: Every().Hour(),
			want:     "0 0 * * * *",
		},
		{
			name:     "every day",
			schedule: Every().Day(),
			want:     "0 0 0 * * *",
		},
		{
			name:     "every 5 seconds",
			schedule: Every().Seconds(5),
			want:     "*/5 * * * * *",
		},
		{
			name:     "every 15 minutes",
			schedule: Every().Minutes(15),
			want:     "0 */15 * * * *",
		},
		{
			name:     "every 6 hours",
			schedule: Every().Hours(6),
			want:     "0 0 */6 * * *",
		},
		{
			name:     "at specific time",
			schedule: Every().Day().At(14, 30),
			want:     "0 30 14 * * *",
		},
		{
			name:     "on specific weekday",
			schedule: Every().Day().OnWeekday(time.Monday),
			want:     "0 0 0 * * 1",
		},
		{
			name:     "on specific day of month",
			schedule: Every().Day().OnDay(15),
			want:     "0 0 0 15 * *",
		},
		{
			name:     "cron expression",
			schedule: Cron("0 0 12 * * *"),
			want:     "0 0 12 * * *",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.schedule.String()
			if got != tt.want {
				t.Errorf("Schedule.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestScheduleBuilderPanics verifies panic conditions
func TestScheduleBuilderPanics(t *testing.T) {
	tests := []struct {
		name string
		fn   func()
	}{
		{
			name: "negative seconds interval",
			fn:   func() { Every().Seconds(-1) },
		},
		{
			name: "zero minutes interval",
			fn:   func() { Every().Minutes(0) },
		},
		{
			name: "invalid hour in At",
			fn:   func() { Every().At(24, 0) },
		},
		{
			name: "invalid minute in At",
			fn:   func() { Every().At(12, 60) },
		},
		{
			name: "invalid day",
			fn:   func() { Every().OnDay(32) },
		},
		{
			name: "zero day",
			fn:   func() { Every().OnDay(0) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Error("Expected panic, but didn't get one")
				}
			}()
			tt.fn()
		})
	}
}

// TestConcurrentOperations verifies thread-safety
func TestConcurrentOperations(t *testing.T) {
	tm := New()
	tm.Start()
	defer tm.Stop()

	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrent task additions
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			task := func(ctx context.Context) error { return nil }
			tm.AddTask(fmt.Sprintf("task-%d", id), Every().Minute(), task)
		}(i)
	}

	wg.Wait()

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tm.ListTasks()
			tm.GetStats()
		}()
	}

	wg.Wait()

	// Concurrent modifications
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			taskName := fmt.Sprintf("task-%d", id)
			tm.DisableTask(taskName)
			tm.EnableTask(taskName)
			tm.GetTask(taskName)
		}(i)
	}

	wg.Wait()
}

// TestTaskInfoCopy verifies that GetTask returns a copy
func TestTaskInfoCopy(t *testing.T) {
	tm := New()
	defer tm.Stop()

	task := func(ctx context.Context) error { return nil }
	err := tm.AddTask("copy-test", Every().Minute(), task)
	if err != nil {
		t.Fatalf("AddTask() failed: %v", err)
	}

	info1, _ := tm.GetTask("copy-test")
	info2, _ := tm.GetTask("copy-test")

	// Modify the copy
	info1.RunCount = 999

	// Original should be unchanged
	if info2.RunCount == 999 {
		t.Error("GetTask() did not return a copy, modifications affected other references")
	}
}

// TestContextCancellation verifies context cancellation handling
func TestContextCancellation(t *testing.T) {
	tm := New()
	tm.Start()

	var taskStarted atomic.Bool
	var taskCompleted atomic.Bool

	task := func(ctx context.Context) error {
		taskStarted.Store(true)
		select {
		case <-time.After(5 * time.Second):
			taskCompleted.Store(true)
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil
	}

	err := tm.AddTask("cancel-test", Every().Second(), task)
	if err != nil {
		t.Fatalf("AddTask() failed: %v", err)
	}

	// Wait for task to start
	time.Sleep(1200 * time.Millisecond)

	if !taskStarted.Load() {
		t.Fatal("Task did not start")
	}

	// Stop immediately
	tm.Stop()

	// Task should not complete normally
	if taskCompleted.Load() {
		t.Error("Task completed normally despite context cancellation")
	}
}

// BenchmarkTaskExecution benchmarks task execution overhead
func BenchmarkTaskExecution(b *testing.B) {
	tm := New()
	tm.Start()
	defer tm.Stop()

	var counter atomic.Int32
	task := func(ctx context.Context) error {
		counter.Add(1)
		return nil
	}

	err := tm.AddTask("bench", Every().Second(), task)
	if err != nil {
		b.Fatalf("AddTask() failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.RunTaskNow("bench")
	}
	b.StopTimer()

	// Wait for all tasks to complete
	time.Sleep(500 * time.Millisecond)
}

// BenchmarkConcurrentTasks benchmarks concurrent task execution
func BenchmarkConcurrentTasks(b *testing.B) {
	tm := New(WithMaxConcurrent(10))
	tm.Start()
	defer tm.Stop()

	task := func(ctx context.Context) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	for i := 0; i < 10; i++ {
		err := tm.AddTask(fmt.Sprintf("bench-%d", i), Every().Second(), task)
		if err != nil {
			b.Fatalf("AddTask() failed: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.RunTaskNow(fmt.Sprintf("bench-%d", i%10))
	}
	b.StopTimer()

	time.Sleep(500 * time.Millisecond)
}
