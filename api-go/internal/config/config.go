package config

import (
	"encoding/json"
	"fmt"
	"strings"
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
	MaxBodySize  int64
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
		MaxBodySize:  env.GetByteSize("HTTP_MAX_BODY_SIZE", 1024*1024*5),
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

func (c *Config) ToJSON(hideSensitive bool) (string, error) {
	// 创建一个副本以避免暴露敏感信息
	safe := *c

	if hideSensitive {
		safe.DB.URL = maskSensitive(safe.DB.URL)
		safe.Redis.URL = maskSensitive(safe.Redis.URL)
		safe.Redis.Password = maskSecret(safe.Redis.Password)
	}

	data, err := json.MarshalIndent(safe, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	return string(data), nil
}

// maskSensitive 遮蔽敏感信息（如密码等）
func maskSensitive(url string) string {
	// 简单的URL密码遮蔽
	if strings.Contains(url, "://") {
		parts := strings.Split(url, "://")
		if len(parts) == 2 {
			scheme := parts[0]
			rest := parts[1]

			// 查找用户信息部分
			if atIndex := strings.Index(rest, "@"); atIndex != -1 {
				userInfo := rest[:atIndex]
				hostPath := rest[atIndex:]

				// 遮蔽密码部分
				if colonIndex := strings.Index(userInfo, ":"); colonIndex != -1 {
					username := userInfo[:colonIndex]
					return fmt.Sprintf("%s://%s:***%s", scheme, username, hostPath)
				}
			}
		}
	}
	return url
}

// maskSecret 遮蔽密钥信息
func maskSecret(secret string) string {
	if secret == "" {
		return ""
	}
	if len(secret) <= 8 {
		return "***"
	}
	return secret[:4] + "***" + secret[len(secret)-4:]
}
