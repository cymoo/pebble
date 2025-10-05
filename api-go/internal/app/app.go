package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	m "github.com/cymoo/mint"
	"github.com/cymoo/pebble/internal/config"
	"github.com/cymoo/pebble/internal/handlers"
	"github.com/cymoo/pebble/internal/services"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
)

type App struct {
	config *config.Config
	db     *sqlx.DB
	redis  *redis.Client
	server *http.Server
}

func New(cfg *config.Config) *App {
	return &App{
		config: cfg,
	}
}

// Initialize 初始化数据库和Redis连接
func (app *App) Initialize() error {
	// 初始化数据库
	if err := app.initDatabase(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	if err := app.initRedis(); err != nil {
		return fmt.Errorf("failed to initialize redis: %w", err)
	}

	// 运行数据库迁移
	// if err := database.Migrate(app.db); err != nil {
	// 	return fmt.Errorf("failed to run migrations: %w", err)
	// }

	// 设置HTTP路由
	app.setupRoutes()

	return nil
}

func (app *App) initDatabase() error {
	db, err := sqlx.Connect("sqlite3", app.config.DB.URL)
	if err != nil {
		log.Printf("Database connection error: %v", app.config.DB.URL)
		return err
	}

	// 配置连接池
	// db.SetMaxOpenConns(app.config.DB.PoolSize) // SQLite 通常只需要 1 个连接
	// db.SetMaxOpenConns(app.config.DB.MaxOpenConns)
	// db.SetMaxIdleConns(app.config.DB.MaxIdleConns)
	// db.SetConnMaxIdleTime(app.config.DB.MaxIdleTime)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return err
	}

	app.db = db
	log.Println("Database connection established successfully")
	return nil
}

func (app *App) initRedis() error {
	app.redis = redis.NewClient(&redis.Options{
		Addr:     app.config.Redis.URL,
		Password: app.config.Redis.Password,
		DB:       app.config.Redis.DB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.redis.Ping(ctx).Err(); err != nil {
		return err
	}

	log.Println("Redis connection established successfully")
	return nil
}

func (app *App) setupRoutes() {
	// 初始化处理器
	// userHandler := handlers.NewUserHandler(app.db, app.redis)

	mux := http.NewServeMux()

	postService := services.NewPostService(app.db)
	postHandler := handlers.NewPostHandler(postService)

	mux.HandleFunc("/hello", m.H(postHandler.HelloWorld))

	// 注册路由
	mux.HandleFunc("GET /health", app.healthHandler)
	// mux.HandleFunc("GET /users/{id}", userHandler.GetUser)

	app.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", app.config.HTTP.IP, app.config.HTTP.Port),
		Handler:      mux,
		ReadTimeout:  app.config.HTTP.ReadTimeout,
		WriteTimeout: app.config.HTTP.WriteTimeout,
		IdleTimeout:  app.config.HTTP.IdleTimeout,
	}
}

func (app *App) healthHandler(w http.ResponseWriter, r *http.Request) {
	// 检查数据库连接
	if err := app.db.Ping(); err != nil {
		http.Error(w, "Database not available", http.StatusServiceUnavailable)
		return
	}

	// 检查Redis连接
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := app.redis.Ping(ctx).Err(); err != nil {
		http.Error(w, "Redis not available", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}

// Run 启动应用服务器
func (app *App) Run() error {
	go func() {
		log.Printf("Server starting on %s", app.server.Addr)
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	return app.Shutdown()
}

// Shutdown 优雅关闭应用
func (app *App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭HTTP服务器
	if err := app.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	// 关闭数据库连接
	if app.db != nil {
		if err := app.db.Close(); err != nil {
			return fmt.Errorf("database connection close failed: %w", err)
		}
	}

	// 关闭Redis连接
	if app.redis != nil {
		if err := app.redis.Close(); err != nil {
			return fmt.Errorf("redis connection close failed: %w", err)
		}
	}

	log.Println("Server shutdown completed")
	return nil
}

// DB 返回数据库实例（用于repository等）
func (app *App) DB() *sqlx.DB {
	return app.db
}

// Redis 返回Redis客户端
func (app *App) Redis() *redis.Client {
	return app.redis
}
