package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cymoo/pebble/pkg/taskmanager"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// App 应用程序依赖（推荐方式）
type App struct {
	Config *Config
	DB     *sqlx.DB
	Redis  *redis.Client
	// 其他依赖...
}

type Config struct {
	AppName string
	Version string
}

var tm *taskmanager.TaskManager

func main() {
	// 初始化应用依赖
	// app := &App{
	// 	Config: &Config{
	// 		AppName: "MyApp",
	// 		Version: "1.0.0",
	// 	},
	// 	DB:    initDB(),
	// 	Redis: initRedis(),
	// }

	// 创建任务管理器，注入依赖
	logger := log.New(os.Stdout, "[TaskManager] ", log.LstdFlags)
	tm = taskmanager.New(
		taskmanager.WithLogger(logger),
		taskmanager.WithLocation(time.Local),
		taskmanager.WithMaxConcurrent(5),
		taskmanager.WithAllowOverlapping(false),
		// taskmanager.WithDependencies(app), // 注入依赖
	)

	// 添加任务
	setupTasks()

	// 启动任务管理器
	tm.Start()

	// 设置 HTTP 路由
	mux := http.NewServeMux()

	// API 端点
	mux.HandleFunc("/api/tasks", handleTasks)
	mux.HandleFunc("/api/tasks/", handleTask)
	mux.HandleFunc("/api/tasks/run/", handleRunTask)
	mux.HandleFunc("/api/tasks/enable/", handleEnableTask)
	mux.HandleFunc("/api/tasks/disable/", handleDisableTask)
	mux.HandleFunc("/api/stats", handleStats)
	mux.HandleFunc("/health", handleHealth)

	// 启动 HTTP 服务器
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		log.Println("Server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// 停止任务管理器
	tm.Stop()

	// 关闭 HTTP 服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// 关闭依赖
	// app.DB.Close()
	// app.Redis.Close()

	log.Println("Server exited")
}

// setupTasks 设置任务（使用依赖注入）
func setupTasks() {
	// 方式 1: 使用 AddTaskWithDeps（推荐）
	// tm.AddTask("clean-data", "0 0 3 * * *", func(ctx context.Context) error {
	tm.AddTask("clean-data", taskmanager.Cron("0 * * * * *"), func(ctx context.Context) error {
		log.Printf("[%s] Cleaning data from database...", "foo")
		time.Sleep(3 * time.Second)

		// 直接使用注入的依赖
		// var count int
		// err := app.DB.GetContext(ctx, &count, "SELECT COUNT(*) FROM old_records WHERE created_at < NOW() - INTERVAL '30 days'")
		// if err != nil {
		// 	return fmt.Errorf("failed to query: %w", err)
		// }

		// log.Printf("Found %d old records to clean", count)

		// // 执行删除
		// _, err = app.DB.ExecContext(ctx, "DELETE FROM old_records WHERE created_at < NOW() - INTERVAL '30 days'")
		// return err
		return nil
	})

	// 方式 2: 使用 Redis 缓存
	// tm.AddTask("clear-cache", "0 0 * * * *", func(ctx context.Context) error {
	tm.AddTask("clear-cache", taskmanager.Cron("0 */2 * * * *"), func(ctx context.Context) error {
		log.Println("Clearing expired cache...")

		// // 使用 Redis
		// keys, err := app.Redis.Keys(ctx, "temp:*").Result()
		// if err != nil {
		// 	return err
		// }

		// if len(keys) > 0 {
		// 	return app.Redis.Del(ctx, keys...).Err()
		// }
		return nil
	})

	// // 方式 3: 健康检查任务
	// tm.AddTaskWithDeps("health-check", "0 */5 * * * *", func(ctx context.Context, app *App) error {
	// 	// 检查数据库
	// 	if err := app.DB.PingContext(ctx); err != nil {
	// 		return fmt.Errorf("database unhealthy: %w", err)
	// 	}

	// 	// 检查 Redis
	// 	if err := app.Redis.Ping(ctx).Err(); err != nil {
	// 		return fmt.Errorf("redis unhealthy: %w", err)
	// 	}

	// 	log.Println("All services healthy")
	// 	return nil
	// })

	// // 方式 4: 生成报告
	// tm.AddTaskWithDeps("generate-report", "0 0 8 * * MON", func(ctx context.Context, app *App) error {
	// 	log.Println("Generating weekly report...")

	// 	type Report struct {
	// 		TotalUsers  int       `db:"total_users"`
	// 		ActiveUsers int       `db:"active_users"`
	// 		Date        time.Time `db:"report_date"`
	// 	}

	// 	var report Report
	// 	query := `
	// 		SELECT
	// 			COUNT(*) as total_users,
	// 			COUNT(*) FILTER (WHERE last_login > NOW() - INTERVAL '7 days') as active_users,
	// 			NOW() as report_date
	// 		FROM users
	// 	`

	// 	if err := app.DB.GetContext(ctx, &report, query); err != nil {
	// 		return err
	// 	}

	// 	// 保存报告到 Redis
	// 	data, _ := json.Marshal(report)
	// 	return app.Redis.Set(ctx, "report:weekly:latest", data, 7*24*time.Hour).Err()
	// })

	// 方式 5: 兼容旧方式（不推荐，但仍支持）
	// 如果你需要通过 context 传递依赖（不推荐）
	// tm.AddTask("legacy-task", "0 0 * * * *", func(ctx context.Context) error {
	// tm.AddTask("legacy-task", "0 */3 * * * *", func(ctx context.Context) error {
	tm.AddTask("legacy-task", taskmanager.Every().Minutes(3), func(ctx context.Context) error {
		// 这种方式需要从 context 中获取依赖，不够优雅
		log.Println("Legacy task executed")
		return nil
	})
}

// // 初始化函数（示例）
// func initDB() *sqlx.DB {
// 	// 实际项目中连接真实数据库
// 	db, _ := sqlx.Open("postgres", "postgresql://localhost/mydb")
// 	return db
// }

// func initRedis() *redis.Client {
// 	return redis.NewClient(&redis.Options{
// 		Addr: "localhost:6379",
// 	})
// }

// HTTP 处理函数

func handleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tasks := tm.ListTasks()

	type TaskResponse struct {
		Name       string    `json:"name"`
		Schedule   string    `json:"schedule"`
		AddedAt    time.Time `json:"added_at"`
		LastRun    time.Time `json:"last_run,omitempty"`
		NextRun    time.Time `json:"next_run,omitempty"`
		RunCount   int64     `json:"run_count"`
		ErrorCount int64     `json:"error_count"`
		LastError  string    `json:"last_error,omitempty"`
		Enabled    bool      `json:"enabled"`
		Running    bool      `json:"running"`
	}

	response := make([]TaskResponse, len(tasks))
	for i, task := range tasks {
		response[i] = TaskResponse{
			Name:       task.Name,
			Schedule:   task.Schedule,
			AddedAt:    task.AddedAt,
			LastRun:    task.LastRun,
			NextRun:    task.NextRun,
			RunCount:   task.RunCount,
			ErrorCount: task.ErrorCount,
			LastError:  task.LastError,
			Enabled:    task.Enabled,
			Running:    task.Running,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleTask(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/api/tasks/"):]
	if name == "" {
		http.Error(w, "Task name required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		task, err := tm.GetTask(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)

	case http.MethodDelete:
		if err := tm.RemoveTask(name); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleRunTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Path[len("/api/tasks/run/"):]
	if name == "" {
		http.Error(w, "Task name required", http.StatusBadRequest)
		return
	}

	if err := tm.RunTaskNow(name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Task '%s' started", name),
	})
}

func handleEnableTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Path[len("/api/tasks/enable/"):]
	if name == "" {
		http.Error(w, "Task name required", http.StatusBadRequest)
		return
	}

	if err := tm.EnableTask(name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Task '%s' enabled", name),
	})
}

func handleDisableTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Path[len("/api/tasks/disable/"):]
	if name == "" {
		http.Error(w, "Task name required", http.StatusBadRequest)
		return
	}

	if err := tm.DisableTask(name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Task '%s' disabled", name),
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	stats := tm.GetStats()
	status := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"stats":     stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := tm.GetStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
