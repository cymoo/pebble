package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cymoo/mita"
	"github.com/cymoo/pebble/internal/config"
	"github.com/cymoo/pebble/internal/tasks"
	"github.com/cymoo/pebble/pkg/fulltext"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
)

type App struct {
	config *config.Config
	db     *sqlx.DB
	redis  *redis.Client
	fts    *fulltext.FullTextSearch
	server *http.Server
	tm     *mita.TaskManager
}

func New(cfg *config.Config) *App {
	return &App{
		config: cfg,
	}
}

// Initialize sets up the application, including database, redis, routes, and tasks
func (app *App) Initialize() error {
	configJSON, err := app.config.ToJSON(true)
	if err != nil {
		panic(err)
	}
	log.Printf("app config:\n%s", configJSON)
	log.Println("=================================")

	if err := app.initDatabase(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// enable foreign key constraints for SQLite
	_, err = app.db.Exec(`PRAGMA foreign_keys = ON;`)
	if err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	if err := app.initRedis(); err != nil {
		return fmt.Errorf("failed to initialize redis: %w", err)
	}

	if err := app.initFullTextSearch(); err != nil {
		return fmt.Errorf("failed to initialize full-text search: %w", err)
	}

	if err := app.setupTasks(); err != nil {
		return fmt.Errorf("failed to add tasks: %w", err)
	}

	app.setupRoutes()

	return nil
}

func (app *App) initDatabase() error {
	db, err := sqlx.Connect("sqlite3", app.config.DB.URL)
	if err != nil {
		log.Printf("database connection error: %v", app.config.DB.URL)
		return err
	}

	verifyForeignKeysConstraints(db)
	verifyWALMode(db)

	// configure the connection pool
	// db.SetMaxOpenConns(app.config.DB.PoolSize) // SQLite 通常只需要 1 个连接
	// db.SetMaxOpenConns(app.config.DB.MaxOpenConns)
	// db.SetMaxIdleConns(app.config.DB.MaxIdleConns)
	// db.SetConnMaxIdleTime(app.config.DB.MaxIdleTime)

	runMigrations(app.config.DB.URL)

	// test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return err
	}

	app.db = db
	log.Println("database connection established successfully")
	return nil
}

func (app *App) initRedis() error {
	app.redis = redis.NewClient(&redis.Options{
		Addr:     app.config.Redis.URL,
		Password: app.config.Redis.Password,
		DB:       app.config.Redis.DB,
	})

	// test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.redis.Ping(ctx).Err(); err != nil {
		return err
	}

	log.Println("redis connection established successfully")
	return nil
}

func (app *App) initFullTextSearch() error {
	app.fts = fulltext.NewFullTextSearch(
		app.redis,
		fulltext.NewGseTokenizer(),
		app.config.Search.PartialMatch,
		app.config.Search.MaxResults,
		app.config.Search.KeyPrefix,
	)
	log.Println("full-text search initialized successfully")
	return nil
}

func (app *App) setupTasks() error {
	tm := mita.New()

	tm.SetContextValue("db", app.db)

	if err := tm.AddTask("delete-old-posts", mita.Every().Day().At(2, 0), tasks.DeleteOldPosts); err != nil {
		return err
	}

	app.tm = tm

	return nil
}

func (app *App) setupRoutes() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(PanicRecovery(app.config.Debug))
	r.Use(CORS(app.config.HTTP.CORS))

	r.Handle(app.config.Upload.BaseURL+"/*", http.StripPrefix(
		app.config.Upload.BaseURL,
		http.FileServer(http.Dir(app.config.Upload.BasePath))),
	)

	r.Get("/health", app.checkHealth)

	// mount task web ui
	r.Mount("/", app.tm.WebHandler("/tasks"))
	r.Mount("/api", NewApiRouter(app))

	app.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", app.config.HTTP.IP, app.config.HTTP.Port),
		Handler:      r,
		ReadTimeout:  app.config.HTTP.ReadTimeout,
		WriteTimeout: app.config.HTTP.WriteTimeout,
		IdleTimeout:  app.config.HTTP.IdleTimeout,
	}
}

func (app *App) checkHealth(w http.ResponseWriter, r *http.Request) {
	// check database connection
	if err := app.db.Ping(); err != nil {
		http.Error(w, "database not available", http.StatusServiceUnavailable)
		return
	}

	// check redis connection
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := app.redis.Ping(ctx).Err(); err != nil {
		http.Error(w, "redis not available", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}

// Run starts the HTTP server and listens for shutdown signals
func (app *App) Run() error {

	// start background tasks
	app.tm.Start()

	go func() {
		log.Printf("server starting on %s", app.server.Addr)
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed to start: %v", err)
		}
	}()

	// wait for interrupt signal to gracefully shutdown the server
	// catch SIGINT and SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	return app.Shutdown()
}

// shutdown cleans up resources and gracefully shuts down the server
func (app *App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// stop background tasks
	app.tm.Stop()

	// gracefully shutdown the server
	if err := app.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	// close database connection
	if app.db != nil {
		if err := app.db.Close(); err != nil {
			return fmt.Errorf("database connection close failed: %w", err)
		}
	}

	// close redis connection
	if app.redis != nil {
		if err := app.redis.Close(); err != nil {
			return fmt.Errorf("redis connection close failed: %w", err)
		}
	}

	log.Println("server shutdown completed")
	return nil
}

func (app *App) GetDB() *sqlx.DB {
	return app.db
}

func (app *App) GetRedis() *redis.Client {
	return app.redis
}

func (app *App) GetFTS() *fulltext.FullTextSearch {
	return app.fts
}

func verifyForeignKeysConstraints(db *sqlx.DB) {
	var rv int
	err := db.Get(&rv, "PRAGMA foreign_keys;")
	if err != nil {
		panic("failed to verify foreign keys constraints: " + err.Error())
	}
	if rv != 1 {
		panic("foreign keys constraints are not enabled")
	}
}

func verifyWALMode(db *sqlx.DB) {
	var rv string
	err := db.Get(&rv, "PRAGMA journal_mode;")
	if err != nil {
		panic("failed to verify WAL mode: " + err.Error())
	}
	if rv != "wal" {
		panic("WAL mode is not enabled")
	}
}

func runMigrations(url string) error {
	migrator, err := migrate.New(
		"file://migrations",
		"sqlite3://"+url,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	err = migrator.Up()
	switch {
	case errors.Is(err, migrate.ErrNoChange):
		log.Println("no new migrations to apply")
		return nil
	case err != nil:
		return fmt.Errorf("migration failed: %w", err)
	default:
		log.Println("migrations applied successfully")
		return nil
	}
}
