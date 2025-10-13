package taskmanager

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// Task 任务函数类型
type Task func(ctx context.Context) error

// TaskInfo 任务信息
type TaskInfo struct {
	Name       string
	Schedule   string
	Task       Task
	EntryID    cron.EntryID
	AddedAt    time.Time
	LastRun    time.Time
	NextRun    time.Time
	RunCount   int64
	ErrorCount int64
	LastError  string
	Enabled    bool
	Running    bool // 标识任务是否正在运行
}

// TaskManager 任务管理器
type TaskManager struct {
	cron             *cron.Cron
	tasks            map[string]*TaskInfo
	mu               sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	logger           *log.Logger
	wg               sync.WaitGroup
	maxConcurrent    int           // 最大并发任务数，0 表示无限制
	semaphore        chan struct{} // 用于限制并发
	allowOverlapping bool          // 是否允许同一任务重叠执行
}

// Option 配置选项
type Option func(*TaskManager)

// WithLogger 设置日志器
func WithLogger(logger *log.Logger) Option {
	return func(tm *TaskManager) {
		tm.logger = logger
	}
}

// WithLocation 设置时区
func WithLocation(loc *time.Location) Option {
	return func(tm *TaskManager) {
		tm.cron = cron.New(cron.WithLocation(loc), cron.WithSeconds())
	}
}

// WithMaxConcurrent 设置最大并发任务数（0 表示无限制）
func WithMaxConcurrent(max int) Option {
	return func(tm *TaskManager) {
		tm.maxConcurrent = max
		if max > 0 {
			tm.semaphore = make(chan struct{}, max)
		}
	}
}

// WithAllowOverlapping 设置是否允许同一任务重叠执行
func WithAllowOverlapping(allow bool) Option {
	return func(tm *TaskManager) {
		tm.allowOverlapping = allow
	}
}

// New 创建新的任务管理器
func New(opts ...Option) *TaskManager {
	ctx, cancel := context.WithCancel(context.Background())

	tm := &TaskManager{
		cron:             cron.New(cron.WithSeconds()), // 支持秒级调度
		tasks:            make(map[string]*TaskInfo),
		ctx:              ctx,
		cancel:           cancel,
		logger:           log.Default(),
		allowOverlapping: false, // 默认不允许重叠
	}

	// 应用选项
	for _, opt := range opts {
		opt(tm)
	}

	return tm
}

// AddTask 添加任务（兼容旧方式）
func (tm *TaskManager) AddTask(name, schedule string, task Task) error {
	return tm.addTaskInternal(name, schedule, func(ctx context.Context) error {
		return task(ctx)
	})
}

// addTaskInternal 内部添加任务的实现
func (tm *TaskManager) addTaskInternal(name, schedule string, task Task) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 检查任务是否已存在
	if _, exists := tm.tasks[name]; exists {
		return fmt.Errorf("task '%s' already exists", name)
	}

	// 包装任务以添加统计和错误处理
	wrappedTask := tm.wrapTask(name, task)

	// 添加到 cron
	entryID, err := tm.cron.AddFunc(schedule, wrappedTask)
	if err != nil {
		return fmt.Errorf("failed to add task '%s': %w", name, err)
	}

	// 保存任务信息
	tm.tasks[name] = &TaskInfo{
		Name:     name,
		Schedule: schedule,
		Task:     task,
		EntryID:  entryID,
		AddedAt:  time.Now(),
		Enabled:  true,
	}

	tm.logger.Printf("Task '%s' added with schedule: %s", name, schedule)
	return nil
}

// wrapTask 包装任务以添加统计和错误处理
func (tm *TaskManager) wrapTask(name string, task Task) func() {
	return func() {
		// 检查任务是否启用
		tm.mu.Lock()
		info := tm.tasks[name]
		if info == nil || !info.Enabled {
			tm.mu.Unlock()
			return
		}

		// 检查是否允许重叠执行
		if !tm.allowOverlapping && info.Running {
			tm.mu.Unlock()
			tm.logger.Printf("Task '%s' is already running, skipping this execution", name)
			return
		}

		info.Running = true
		info.LastRun = time.Now()
		info.RunCount++
		tm.mu.Unlock()

		// 限流：如果设置了最大并发数
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

		tm.wg.Add(1)
		defer func() {
			tm.wg.Done()
			tm.mu.Lock()
			info.Running = false
			tm.mu.Unlock()
		}()

		// 执行任务
		startTime := time.Now()
		if err := task(tm.ctx); err != nil {
			tm.mu.Lock()
			info.ErrorCount++
			info.LastError = err.Error()
			tm.mu.Unlock()
			tm.logger.Printf("Task '%s' failed after %v: %v", name, time.Since(startTime), err)
		} else {
			tm.mu.Lock()
			info.LastError = "" // 清除之前的错误
			tm.mu.Unlock()
			tm.logger.Printf("Task '%s' completed successfully in %v", name, time.Since(startTime))
		}

		// 更新下次运行时间
		tm.updateNextRun(name)
	}
}

// updateNextRun 更新下次运行时间
func (tm *TaskManager) updateNextRun(name string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if info, ok := tm.tasks[name]; ok {
		entry := tm.cron.Entry(info.EntryID)
		info.NextRun = entry.Next
	}
}

// RemoveTask 移除任务
func (tm *TaskManager) RemoveTask(name string) error {
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

// EnableTask 启用任务
func (tm *TaskManager) EnableTask(name string) error {
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

// DisableTask 禁用任务
func (tm *TaskManager) DisableTask(name string) error {
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

// GetTask 获取任务信息
func (tm *TaskManager) GetTask(name string) (*TaskInfo, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	info, exists := tm.tasks[name]
	if !exists {
		return nil, fmt.Errorf("task '%s' not found", name)
	}

	// 更新下次运行时间
	entry := tm.cron.Entry(info.EntryID)
	infoCopy := *info
	infoCopy.NextRun = entry.Next

	return &infoCopy, nil
}

// ListTasks 列出所有任务
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

// Start 启动任务管理器
func (tm *TaskManager) Start() {
	tm.cron.Start()
	tm.logger.Println("Task manager started")
}

// Stop 停止任务管理器
func (tm *TaskManager) Stop() {
	tm.logger.Println("Stopping task manager...")
	tm.cancel()

	// 停止 cron 调度器
	ctx := tm.cron.Stop()

	// 等待所有正在运行的任务完成
	done := make(chan struct{})
	go func() {
		tm.wg.Wait()
		close(done)
	}()

	// 等待任务完成或超时
	select {
	case <-done:
		tm.logger.Println("All tasks completed")
	case <-ctx.Done():
		tm.logger.Println("Task manager stopped")
	case <-time.After(30 * time.Second):
		tm.logger.Println("Timeout waiting for tasks to complete")
	}
}

// RunTaskNow 立即运行指定任务
func (tm *TaskManager) RunTaskNow(name string) error {
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

	// 检查是否允许重叠执行
	if !tm.allowOverlapping && info.Running {
		tm.mu.Unlock()
		return fmt.Errorf("task '%s' is already running", name)
	}

	info.Running = true
	info.LastRun = time.Now()
	info.RunCount++
	task := info.Task
	tm.mu.Unlock()

	// 异步执行任务
	go func() {
		// 限流：如果设置了最大并发数
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

		tm.wg.Add(1)
		defer func() {
			tm.wg.Done()
			tm.mu.Lock()
			info.Running = false
			tm.mu.Unlock()
		}()

		startTime := time.Now()
		if err := task(tm.ctx); err != nil {
			tm.mu.Lock()
			info.ErrorCount++
			info.LastError = err.Error()
			tm.mu.Unlock()
			tm.logger.Printf("Manual run of task '%s' failed after %v: %v", name, time.Since(startTime), err)
		} else {
			tm.mu.Lock()
			info.LastError = ""
			tm.mu.Unlock()
			tm.logger.Printf("Manual run of task '%s' completed successfully in %v", name, time.Since(startTime))
		}
	}()

	return nil
}

// GetStats 获取任务管理器统计信息
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

// IsRunning 检查任务管理器是否正在运行
func (tm *TaskManager) IsRunning() bool {
	select {
	case <-tm.ctx.Done():
		return false
	default:
		return true
	}
}
