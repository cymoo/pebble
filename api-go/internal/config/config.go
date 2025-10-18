package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cymoo/pebble/pkg/env"
)

type Config struct {
	// Basic app info
	AppName    string
	AppVersion string
	AppEnv     string
	Debug      bool

	// Application settings
	PostsPerPage int
	StaticURL    string
	StaticPath   string

	// Server settings
	HTTP   HTTPConfig
	Upload UploadConfig

	DB    DBConfig
	Redis RedisConfig
}

type UploadConfig struct {
	BaseURL      string
	BasePath     string
	ImageFormats []string
	ThumbWidth   uint32
}

type DBConfig struct {
	URL         string
	PoolSize    int
	AutoMigrate bool
}

type RedisConfig struct {
	URL      string
	Password string
	DB       int
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
	config.AppEnv = envType
	config.Debug = envType == "development" || envType == "dev"

	env.LoadConfigFiles(envType)

	config.AppName = env.GetString("APP_NAME", "Pebble")
	config.AppVersion = env.GetString("APP_VERSION", "1.0.0")

	config.PostsPerPage = env.GetInt("POSTS_PER_PAGE", 30)

	config.StaticURL = env.GetString("STATIC_URL", "/static")
	// If StaticPath is not set, then static files will be served from embedded FS
	config.StaticPath = env.GetString("STATIC_PATH", "")

	config.HTTP = HTTPConfig{
		IP:           env.GetString("HTTP_IP", "localhost"),
		Port:         env.GetInt("HTTP_PORT", 8000),
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

	config.Upload = UploadConfig{
		BaseURL:      env.GetString("UPLOAD_URL", "/uploads"),
		BasePath:     env.GetString("UPLOAD_PATH", "./uploads"),
		ImageFormats: env.GetSlice("UPLOAD_IMAGE_FORMATS", []string{"jpg", "jpeg", "png", "webp", "gif"}),
		ThumbWidth:   uint32(env.GetInt("UPLOAD_THUMB_WIDTH", 128)),
	}

	config.DB = DBConfig{
		URL:         env.GetString("DB_URL", "app.db"),
		PoolSize:    env.GetInt("DB_POOL_SIZE", 5),
		AutoMigrate: env.GetBool("DB_AUTO_MIGRATE", true),
	}

	config.Redis = RedisConfig{
		URL:      env.GetString("REDIS_URL", "localhost:6379"),
		Password: env.GetString("REDIS_PASSWORD", ""),
		DB:       env.GetInt("REDIS_DB", 0),
	}

	return config
}

func (c *Config) ToJSON(hideSensitive bool) (string, error) {
	// Create a copy to avoid exposing sensitive info
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

// maskSensitive masks sensitive information in URLs
func maskSensitive(url string) string {
	// Check if it contains "://"
	if strings.Contains(url, "://") {
		parts := strings.Split(url, "://")
		if len(parts) == 2 {
			scheme := parts[0]
			rest := parts[1]

			// Look for user info part
			if atIndex := strings.Index(rest, "@"); atIndex != -1 {
				userInfo := rest[:atIndex]
				hostPath := rest[atIndex:]

				// Mask password part
				if colonIndex := strings.Index(userInfo, ":"); colonIndex != -1 {
					username := userInfo[:colonIndex]
					return fmt.Sprintf("%s://%s:***%s", scheme, username, hostPath)
				}
			}
		}
	}
	return url
}

// maskSecret masks a secret string, showing only the first and last 4 characters
func maskSecret(secret string) string {
	if secret == "" {
		return ""
	}
	if len(secret) <= 8 {
		return "***"
	}
	return secret[:4] + "***" + secret[len(secret)-4:]
}
