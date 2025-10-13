package taskmanager

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
	// taskNameKey is the context key for storing the task name.
	taskNameKey contextKey = "taskName"
)

// Task represents a function that performs work within a given context.
// It should return an error if the task execution fails.
type Task func(ctx context.Context) error

// TaskInfo holds metadata and statistics about a scheduled task.
type TaskInfo struct {
	Name       string       // Unique identifier for the task
	Schedule   string       // Cron expression for the task schedule
	Task       Task         // The actual task function to execute
	EntryID    cron.EntryID // Cron entry ID for this task
	AddedAt    time.Time    // When the task was added to the manager
	LastRun    time.Time    // Last execution time
	NextRun    time.Time    // Next scheduled execution time
	RunCount   int64        // Total number of executions
	ErrorCount int64        // Total number of failed executions
	LastError  string       // Most recent error message (empty if last run succeeded)
	Enabled    bool         // Whether the task is enabled for execution
	Running    bool         // Whether the task is currently executing
}

// TaskManager orchestrates scheduled task execution with concurrent control,
// error tracking, and flexible configuration options.
type TaskManager struct {
	cron             *cron.Cron                                                 // Underlying cron scheduler
	tasks            map[string]*TaskInfo                                       // Map of task name to task info
	mu               sync.RWMutex                                               // Protects tasks map and task info
	ctx              context.Context                                            // Manager lifecycle context
	cancel           context.CancelFunc                                         // Function to cancel the manager context
	logger           *log.Logger                                                // Logger for task execution events
	wg               sync.WaitGroup                                             // Tracks running tasks for graceful shutdown
	maxConcurrent    int                                                        // Maximum concurrent tasks (0 = unlimited)
	semaphore        chan struct{}                                              // Channel-based semaphore for concurrency control
	allowOverlapping bool                                                       // Whether same task can run concurrently
	contextValues    map[string]any                                             // Static values to inject into task contexts
	contextInjector  func(ctx context.Context, taskName string) context.Context // Dynamic context injection
}

// Option is a functional option for configuring TaskManager.
type Option func(*TaskManager)

// WithLogger sets a custom logger for the task manager.
// If not provided, log.Default() will be used.
func WithLogger(logger *log.Logger) Option {
	return func(tm *TaskManager) {
		if logger != nil {
			tm.logger = logger
		}
	}
}

// WithLocation sets the timezone for cron schedule interpretation.
func WithLocation(loc *time.Location) Option {
	return func(tm *TaskManager) {
		if loc != nil {
			tm.cron = cron.New(cron.WithLocation(loc), cron.WithSeconds())
		}
	}
}

// WithMaxConcurrent sets the maximum number of tasks that can run concurrently.
// A value of 0 means unlimited concurrency. Negative values are treated as 0.
func WithMaxConcurrent(max int) Option {
	return func(tm *TaskManager) {
		if max < 0 {
			max = 0
		}
		tm.maxConcurrent = max
		if max > 0 {
			tm.semaphore = make(chan struct{}, max)
		}
	}
}

// WithAllowOverlapping controls whether the same task can run multiple instances concurrently.
// By default, overlapping is not allowed (false).
func WithAllowOverlapping(allow bool) Option {
	return func(tm *TaskManager) {
		tm.allowOverlapping = allow
	}
}

// WithContextValue adds a static key-value pair that will be injected into all task contexts.
func WithContextValue(key string, value any) Option {
	return func(tm *TaskManager) {
		if key == "" {
			return
		}
		if tm.contextValues == nil {
			tm.contextValues = make(map[string]any)
		}
		tm.contextValues[key] = value
	}
}

// WithContextInjector sets a custom function to dynamically inject values into task contexts.
// The injector is called for each task execution and receives the base context and task name.
func WithContextInjector(injector func(ctx context.Context, taskName string) context.Context) Option {
	return func(tm *TaskManager) {
		tm.contextInjector = injector
	}
}

// New creates a new TaskManager with the given options.
// The manager must be started with Start() before tasks will execute.
func New(opts ...Option) *TaskManager {
	ctx, cancel := context.WithCancel(context.Background())

	tm := &TaskManager{
		cron:             cron.New(cron.WithSeconds()), // Support second-level scheduling
		tasks:            make(map[string]*TaskInfo),
		ctx:              ctx,
		cancel:           cancel,
		logger:           log.Default(),
		allowOverlapping: false, // Default: prevent overlapping executions
	}

	// Apply functional options
	for _, opt := range opts {
		opt(tm)
	}

	return tm
}

// AddTask registers a new task with the given name and schedule.
// Returns an error if a task with the same name already exists or if the schedule is invalid.
func (tm *TaskManager) AddTask(name string, schedule Schedule, task Task) error {
	if name == "" {
		return fmt.Errorf("task name cannot be empty")
	}
	if schedule == nil {
		return fmt.Errorf("schedule cannot be nil")
	}
	if task == nil {
		return fmt.Errorf("task function cannot be nil")
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Check if task already exists
	if _, exists := tm.tasks[name]; exists {
		return fmt.Errorf("task '%s' already exists", name)
	}

	// Wrap task to add statistics and error handling
	wrappedTask := tm.wrapTask(name, task)

	// Add to cron scheduler
	entryID, err := tm.cron.AddFunc(schedule.String(), wrappedTask)
	if err != nil {
		return fmt.Errorf("failed to add task '%s': %w", name, err)
	}

	// Store task information
	tm.tasks[name] = &TaskInfo{
		Name:     name,
		Schedule: schedule.String(),
		Task:     task,
		EntryID:  entryID,
		AddedAt:  time.Now(),
		Enabled:  true,
	}

	tm.logger.Printf("Task '%s' added with schedule: %s", name, schedule)
	return nil
}

// wrapTask wraps a task function with execution tracking, error handling,
// concurrency control, and overlap prevention.
func (tm *TaskManager) wrapTask(name string, task Task) func() {
	return func() {
		// Check if task is enabled and handle overlap prevention
		tm.mu.Lock()
		info := tm.tasks[name]
		if info == nil || !info.Enabled {
			tm.mu.Unlock()
			return
		}

		// Prevent overlapping executions if configured
		if !tm.allowOverlapping && info.Running {
			tm.mu.Unlock()
			tm.logger.Printf("Task '%s' is already running, skipping this execution", name)
			return
		}

		info.Running = true
		info.LastRun = time.Now()
		info.RunCount++
		tm.mu.Unlock()

		// Acquire semaphore if concurrency limit is set
		if tm.semaphore != nil {
			select {
			case tm.semaphore <- struct{}{}:
				defer func() { <-tm.semaphore }()
			case <-tm.ctx.Done():
				tm.mu.Lock()
				info.Running = false
				tm.mu.Unlock()
				return
			}
		}

		// Track execution for graceful shutdown
		tm.wg.Add(1)
		defer func() {
			tm.wg.Done()
			tm.mu.Lock()
			info.Running = false
			tm.mu.Unlock()
		}()

		// Execute the task
		startTime := time.Now()
		if err := task(tm.getTaskContext(name)); err != nil {
			tm.mu.Lock()
			info.ErrorCount++
			info.LastError = err.Error()
			tm.mu.Unlock()
			tm.logger.Printf("Task '%s' failed after %v: %v", name, time.Since(startTime), err)
		} else {
			tm.mu.Lock()
			info.LastError = "" // Clear previous error on success
			tm.mu.Unlock()
			tm.logger.Printf("Task '%s' completed successfully in %v", name, time.Since(startTime))
		}

		// Update next run time
		tm.updateNextRun(name)
	}
}

// updateNextRun updates the next scheduled run time for a task.
func (tm *TaskManager) updateNextRun(name string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if info, ok := tm.tasks[name]; ok {
		entry := tm.cron.Entry(info.EntryID)
		info.NextRun = entry.Next
	}
}

// RemoveTask removes a task from the manager.
// Returns an error if the task does not exist.
func (tm *TaskManager) RemoveTask(name string) error {
	if name == "" {
		return fmt.Errorf("task name cannot be empty")
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	info, exists := tm.tasks[name]
	if !exists {
		return fmt.Errorf("task '%s' not found", name)
	}

	tm.cron.Remove(info.EntryID)
	delete(tm.tasks, name)
	tm.logger.Printf("Task '%s' removed", name)
	return nil
}

// EnableTask enables a previously disabled task.
// The task will resume executing on its schedule.
func (tm *TaskManager) EnableTask(name string) error {
	if name == "" {
		return fmt.Errorf("task name cannot be empty")
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	info, exists := tm.tasks[name]
	if !exists {
		return fmt.Errorf("task '%s' not found", name)
	}

	info.Enabled = true
	tm.logger.Printf("Task '%s' enabled", name)
	return nil
}

// DisableTask disables a task without removing it.
// The task will not execute but can be re-enabled later.
func (tm *TaskManager) DisableTask(name string) error {
	if name == "" {
		return fmt.Errorf("task name cannot be empty")
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	info, exists := tm.tasks[name]
	if !exists {
		return fmt.Errorf("task '%s' not found", name)
	}

	info.Enabled = false
	tm.logger.Printf("Task '%s' disabled", name)
	return nil
}

// GetTask returns a copy of the task information for the given task name.
// Returns an error if the task does not exist.
func (tm *TaskManager) GetTask(name string) (*TaskInfo, error) {
	if name == "" {
		return nil, fmt.Errorf("task name cannot be empty")
	}

	tm.mu.RLock()
	defer tm.mu.RUnlock()

	info, exists := tm.tasks[name]
	if !exists {
		return nil, fmt.Errorf("task '%s' not found", name)
	}

	// Update next run time and return a copy
	entry := tm.cron.Entry(info.EntryID)
	infoCopy := *info
	infoCopy.NextRun = entry.Next

	return &infoCopy, nil
}

// ListTasks returns a copy of all task information.
// The returned slice can be safely modified without affecting the manager.
func (tm *TaskManager) ListTasks() []*TaskInfo {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tasks := make([]*TaskInfo, 0, len(tm.tasks))
	for _, info := range tm.tasks {
		entry := tm.cron.Entry(info.EntryID)
		infoCopy := *info
		infoCopy.NextRun = entry.Next
		tasks = append(tasks, &infoCopy)
	}

	return tasks
}

// Start begins the task scheduler.
// Tasks will start executing according to their schedules.
func (tm *TaskManager) Start() {
	tm.cron.Start()
	tm.logger.Println("Task manager started")
}

// Stop gracefully shuts down the task manager.
// It stops accepting new task executions and waits for running tasks to complete
// or times out after 30 seconds.
func (tm *TaskManager) Stop() {
	tm.logger.Println("Stopping task manager...")

	// Cancel the context to stop new executions
	tm.cancel()

	// Stop the cron scheduler and get its context
	cronCtx := tm.cron.Stop()

	// Wait for all running tasks to complete
	done := make(chan struct{})
	go func() {
		tm.wg.Wait()
		close(done)
	}()

	// Wait for completion or timeout
	select {
	case <-done:
		tm.logger.Println("All tasks completed gracefully")
	case <-cronCtx.Done():
		// Cron stopped, still wait a bit for tasks
		select {
		case <-done:
			tm.logger.Println("All tasks completed after cron stop")
		case <-time.After(30 * time.Second):
			tm.logger.Println("Timeout waiting for tasks to complete")
		}
	case <-time.After(30 * time.Second):
		tm.logger.Println("Timeout waiting for tasks to complete")
	}
}

// RunTaskNow immediately executes a task outside of its regular schedule.
// The execution is asynchronous and subject to the same concurrency limits
// and overlap rules as scheduled executions.
func (tm *TaskManager) RunTaskNow(name string) error {
	if name == "" {
		return fmt.Errorf("task name cannot be empty")
	}

	tm.mu.Lock()
	info, exists := tm.tasks[name]
	if !exists {
		tm.mu.Unlock()
		return fmt.Errorf("task '%s' not found", name)
	}

	if !info.Enabled {
		tm.mu.Unlock()
		return fmt.Errorf("task '%s' is disabled", name)
	}

	// Check overlap prevention
	if !tm.allowOverlapping && info.Running {
		tm.mu.Unlock()
		return fmt.Errorf("task '%s' is already running", name)
	}

	info.Running = true
	info.LastRun = time.Now()
	info.RunCount++
	task := info.Task
	tm.mu.Unlock()

	// Execute task asynchronously
	go func() {
		// Acquire semaphore if concurrency limit is set
		if tm.semaphore != nil {
			select {
			case tm.semaphore <- struct{}{}:
				defer func() { <-tm.semaphore }()
			case <-tm.ctx.Done():
				tm.mu.Lock()
				if taskInfo, ok := tm.tasks[name]; ok {
					taskInfo.Running = false
				}
				tm.mu.Unlock()
				return
			}
		}

		tm.wg.Add(1)
		defer func() {
			tm.wg.Done()
			tm.mu.Lock()
			if taskInfo, ok := tm.tasks[name]; ok {
				taskInfo.Running = false
			}
			tm.mu.Unlock()
		}()

		startTime := time.Now()
		if err := task(tm.getTaskContext(name)); err != nil {
			tm.mu.Lock()
			if taskInfo, ok := tm.tasks[name]; ok {
				taskInfo.ErrorCount++
				taskInfo.LastError = err.Error()
			}
			tm.mu.Unlock()
			tm.logger.Printf("Manual run of task '%s' failed after %v: %v", name, time.Since(startTime), err)
		} else {
			tm.mu.Lock()
			if taskInfo, ok := tm.tasks[name]; ok {
				taskInfo.LastError = ""
			}
			tm.mu.Unlock()
			tm.logger.Printf("Manual run of task '%s' completed successfully in %v", name, time.Since(startTime))
		}
	}()

	return nil
}

// SetContextValue adds or updates a static context value that will be
// injected into all task contexts.
func (tm *TaskManager) SetContextValue(key string, value any) {
	if key == "" {
		return
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.contextValues == nil {
		tm.contextValues = make(map[string]any)
	}
	tm.contextValues[key] = value
}

// GetContextValue retrieves a static context value.
// Returns nil if the key does not exist.
func (tm *TaskManager) GetContextValue(key string) any {
	if key == "" {
		return nil
	}

	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if tm.contextValues == nil {
		return nil
	}
	return tm.contextValues[key]
}

// getTaskContext creates a context for task execution with injected values.
// It includes the task name, static context values, and any custom injections.
func (tm *TaskManager) getTaskContext(name string) context.Context {
	ctx := tm.ctx

	// Inject task name using typed key
	ctx = context.WithValue(ctx, taskNameKey, name)

	// Inject static context values (copy to avoid holding lock during injection)
	var contextValues map[string]any
	if len(tm.contextValues) > 0 {
		tm.mu.RLock()
		contextValues = make(map[string]any, len(tm.contextValues))
		for key, value := range tm.contextValues {
			contextValues[key] = value
		}
		tm.mu.RUnlock()
	}

	// Apply static values without holding lock
	for key, value := range contextValues {
		ctx = context.WithValue(ctx, key, value)
	}

	// Call custom injector without holding any locks
	if tm.contextInjector != nil {
		ctx = tm.contextInjector(ctx, name)
	}

	return ctx
}

// GetTaskName extracts the task name from a task context.
// Returns empty string if the context doesn't contain a task name.
func GetTaskName(ctx context.Context) string {
	if name, ok := ctx.Value(taskNameKey).(string); ok {
		return name
	}
	return ""
}

// GetStats returns aggregated statistics about the task manager and all tasks.
func (tm *TaskManager) GetStats() map[string]interface{} {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	totalTasks := len(tm.tasks)
	enabledTasks := 0
	runningTasks := 0
	totalRuns := int64(0)
	totalErrors := int64(0)

	for _, info := range tm.tasks {
		if info.Enabled {
			enabledTasks++
		}
		if info.Running {
			runningTasks++
		}
		totalRuns += info.RunCount
		totalErrors += info.ErrorCount
	}

	return map[string]interface{}{
		"total_tasks":       totalTasks,
		"enabled_tasks":     enabledTasks,
		"running_tasks":     runningTasks,
		"total_runs":        totalRuns,
		"total_errors":      totalErrors,
		"max_concurrent":    tm.maxConcurrent,
		"allow_overlapping": tm.allowOverlapping,
	}
}

// IsRunning checks whether the task manager is currently running.
func (tm *TaskManager) IsRunning() bool {
	select {
	case <-tm.ctx.Done():
		return false
	default:
		return true
	}
}

// Schedule represents a task scheduling expression.
type Schedule interface {
	String() string
}

// CronSchedule wraps a standard cron expression string.
type CronSchedule struct {
	expression string
}

// Cron creates a Schedule from a cron expression string.
// The expression should be in the format: "second minute hour day month weekday"
func Cron(expr string) *CronSchedule {
	return &CronSchedule{expression: expr}
}

// String returns the cron expression.
func (c *CronSchedule) String() string {
	return c.expression
}

// ScheduleBuilder provides a fluent API for building cron schedules.
type ScheduleBuilder struct {
	second  string
	minute  string
	hour    string
	day     string
	month   string
	weekday string
}

// String converts the builder to a cron expression string.
func (s *ScheduleBuilder) String() string {
	return fmt.Sprintf("%s %s %s %s %s %s",
		s.second, s.minute, s.hour, s.day, s.month, s.weekday)
}

// Every creates a new ScheduleBuilder with default values (runs every minute).
func Every() *ScheduleBuilder {
	return &ScheduleBuilder{
		second:  "0",
		minute:  "*",
		hour:    "*",
		day:     "*",
		month:   "*",
		weekday: "*",
	}
}

// Second configures the schedule to run every second.
func (s *ScheduleBuilder) Second() *ScheduleBuilder {
	s.second = "*"
	s.minute = "*"
	s.hour = "*"
	return s
}

// Minute configures the schedule to run every minute (at 0 seconds).
func (s *ScheduleBuilder) Minute() *ScheduleBuilder {
	s.second = "0"
	s.minute = "*"
	s.hour = "*"
	return s
}

// Hour configures the schedule to run every hour (at 0 minutes, 0 seconds).
func (s *ScheduleBuilder) Hour() *ScheduleBuilder {
	s.second = "0"
	s.minute = "0"
	s.hour = "*"
	return s
}

// Day configures the schedule to run every day (at midnight).
func (s *ScheduleBuilder) Day() *ScheduleBuilder {
	s.second = "0"
	s.minute = "0"
	s.hour = "0"
	s.day = "*"
	return s
}

// Seconds configures the schedule to run at the specified second interval.
// For example, Seconds(30) runs every 30 seconds.
// Interval must be positive, otherwise the method panics.
func (s *ScheduleBuilder) Seconds(interval int) *ScheduleBuilder {
	if interval <= 0 {
		panic("interval must be positive")
	}
	s.second = fmt.Sprintf("*/%d", interval)
	s.minute = "*"
	s.hour = "*"
	return s
}

// Minutes configures the schedule to run at the specified minute interval.
// For example, Minutes(15) runs every 15 minutes.
// Interval must be positive, otherwise the method panics.
func (s *ScheduleBuilder) Minutes(interval int) *ScheduleBuilder {
	if interval <= 0 {
		panic("interval must be positive")
	}
	s.second = "0"
	s.minute = fmt.Sprintf("*/%d", interval)
	s.hour = "*"
	return s
}

// Hours configures the schedule to run at the specified hour interval.
// For example, Hours(6) runs every 6 hours.
// Interval must be positive, otherwise the method panics.
func (s *ScheduleBuilder) Hours(interval int) *ScheduleBuilder {
	if interval <= 0 {
		panic("interval must be positive")
	}
	s.second = "0"
	s.minute = "0"
	s.hour = fmt.Sprintf("*/%d", interval)
	return s
}

// Days configures the schedule to run at the specified day interval.
// For example, Days(2) runs every 2 days at midnight.
// Interval must be positive, otherwise the method panics.
func (s *ScheduleBuilder) Days(interval int) *ScheduleBuilder {
	if interval <= 0 {
		panic("interval must be positive")
	}
	s.second = "0"
	s.minute = "0"
	s.hour = "0"
	s.day = fmt.Sprintf("*/%d", interval)
	return s
}

// At specifies a specific time of day for the schedule.
// For example, At(14, 30) runs at 2:30 PM.
// Hour must be 0-23 and minute must be 0-59, otherwise the method panics.
func (s *ScheduleBuilder) At(hour, minute int) *ScheduleBuilder {
	if hour < 0 || hour > 23 {
		panic("hour must be between 0 and 23")
	}
	if minute < 0 || minute > 59 {
		panic("minute must be between 0 and 59")
	}
	s.second = "0"
	s.minute = fmt.Sprintf("%d", minute)
	s.hour = fmt.Sprintf("%d", hour)
	return s
}

// OnWeekday restricts the schedule to a specific day of the week.
// For example, OnWeekday(time.Monday) runs only on Mondays.
func (s *ScheduleBuilder) OnWeekday(weekday time.Weekday) *ScheduleBuilder {
	s.weekday = fmt.Sprintf("%d", weekday)
	return s
}

// OnDay restricts the schedule to a specific day of the month.
// For example, OnDay(15) runs on the 15th of each month.
// Day must be between 1 and 31, otherwise the method panics.
func (s *ScheduleBuilder) OnDay(day int) *ScheduleBuilder {
	if day < 1 || day > 31 {
		panic("day must be between 1 and 31")
	}
	s.day = fmt.Sprintf("%d", day)
	return s
}
