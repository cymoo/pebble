package config

import (
	"time"

	"github.com/cymoo/pebble/pkg/env"
)

type Config struct {
	// 各模块配置
	HTTP   HTTPConfig
	Upload UploadConfig
	Search SearchConfig
	DB     DBConfig
	Redis  RedisConfig

	// 基本信息
	AppName     string
	AppVersion  string
	Environment string
	Debug       bool

	// 业务配置
	PostsPerPage int
	StaticDir    string
	StaticURL    string
}

type UploadConfig struct {
	MaxSize      int64
	BaseURL      string
	BasePath     string
	ImageFormats []string
	ThumbWidth   uint32
}

type SearchConfig struct {
	MaxResults   int
	PartialMatch bool
	KeyPrefix    string
}

type DBConfig struct {
	URL      string
	PoolSize int
}

type RedisConfig struct {
	URL      string
	Password string
	DB       int
	PoolSize int
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

type HTTPConfig struct {
	IP           string
	Port         int
	MaxBodySize  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	CORS         CORSConfig
}

func Load() *Config {
	config := &Config{}
	envType := env.GetString("APP_ENV", "development")
	config.Environment = envType
	config.Debug = envType == "development" || envType == "dev"

	env.LoadConfigFiles(envType)

	config.HTTP = HTTPConfig{
		IP:           env.GetString("HTTP_IP", "localhost"),
		Port:         env.GetInt("HTTP_PORT", 8080),
		ReadTimeout:  env.GetDuration("HTTP_READ_TIMEOUT", 10*time.Second),
		WriteTimeout: env.GetDuration("HTTP_WRITE_TIMEOUT", 10*time.Second),
		IdleTimeout:  env.GetDuration("HTTP_IDLE_TIMEOUT", 30*time.Second),
		CORS: CORSConfig{
			AllowedOrigins:   env.GetSlice("CORS_ALLOWED_ORIGINS", []string{}),
			AllowedMethods:   env.GetSlice("CORS_ALLOWED_METHODS", []string{}),
			AllowedHeaders:   env.GetSlice("CORS_ALLOWED_HEADERS", []string{}),
			AllowCredentials: env.GetBool("CORS_ALLOW_CREDENTIALS", true),
			MaxAge:           env.GetInt("CORS_MAX_AGE", 3600*24),
		},
	}

	config.DB = DBConfig{
		URL:      env.GetString("DB_URL", "file:app.db?cache=shared&_fk=true&_journal_mode=WAL"),
		PoolSize: env.GetInt("DB_POOL_SIZE", 10),
	}

	config.Redis = RedisConfig{
		URL:      env.GetString("REDIS_URL", "localhost:6379"),
		Password: env.GetString("REDIS_PASSWORD", ""),
		DB:       env.GetInt("REDIS_DB", 0),
		PoolSize: env.GetInt("REDIS_POOL_SIZE", 10),
	}

	config.Upload = UploadConfig{
		// MaxSize:    env.GetByteSize("UPLOAD_MAX_SIZE", 10*1024*1024), // 10 MB
		BaseURL:      env.GetString("UPLOAD_BASE_URL", "/uploads/"),
		BasePath:     env.GetString("UPLOAD_BASE_PATH", "./uploads"),
		ImageFormats: env.GetSlice("UPLOAD_IMAGE_FORMATS", []string{"jpg", "jpeg", "png", "webp", "gif"}),
		ThumbWidth:   uint32(env.GetInt("UPLOAD_THUMB_WIDTH", 200)),
	}

	config.Search = SearchConfig{
		MaxResults:   env.GetInt("SEARCH_MAX_RESULTS", 100),
		PartialMatch: env.GetBool("SEARCH_PARTIAL_MATCH", true),
		KeyPrefix:    env.GetString("SEARCH_KEY_PREFIX", ""),
	}

	config.AppName = env.GetString("APP_NAME", "Pebble")
	config.AppVersion = env.GetString("APP_VERSION", "0.1.0")

	config.PostsPerPage = env.GetInt("POSTS_PER_PAGE", 30)
	config.StaticDir = env.GetString("STATIC_PATH", "./static")
	config.StaticURL = env.GetString("STATIC_BASE_URL", "/static/")

	return config
}
