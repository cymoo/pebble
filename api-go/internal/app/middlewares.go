package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/cymoo/pebble/internal/config"
	e "github.com/cymoo/pebble/internal/errors"
	"github.com/cymoo/pebble/internal/services"
	"github.com/redis/go-redis/v9"
)

func PanicRecovery(logTrace bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("panic recovered: %v\n", err)
					if logTrace {
						log.Printf("stack trace:\n%s", debug.Stack())
					}
					e.SendJSONError(w, 500, "internal_error")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// CORS 中间件
// CORSOrigins: 允许的来源列表，如 []string{"http://localhost:3000", "https://example.com"}
// 如果为空或包含 "*"，则允许所有来源
func CORS(config config.CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// 检查并设置允许的Origin
			if len(config.AllowedOrigins) == 0 {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				for _, allowedOrigin := range config.AllowedOrigins {
					if allowedOrigin == "*" || allowedOrigin == origin {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}

			// 设置允许的方法
			methods := "GET, POST, PUT, DELETE, OPTIONS"
			if len(config.AllowedMethods) > 0 {
				methods = strings.Join(config.AllowedMethods, ", ")
			}
			w.Header().Set("Access-Control-Allow-Methods", methods)

			// 设置允许的头部
			headers := "Content-Type, Authorization"
			if len(config.AllowedHeaders) > 0 {
				headers = strings.Join(config.AllowedHeaders, ", ")
			}
			w.Header().Set("Access-Control-Allow-Headers", headers)

			// 设置是否允许凭据
			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// 设置预检请求缓存时间
			if config.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
			}

			// 处理预检请求
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimit returns a net/http middleware that enforces rate limiting
func RateLimit(client *redis.Client, expires time.Duration, maxCount int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := fmt.Sprintf("rate:%s", r.URL.Path)

			belowLimit, err := checkRateLimit(r.Context(), client, key, expires, maxCount)
			if err != nil {
				log.Printf("error checking rate limit: %v", err)
				e.SendJSONError(w, 500, "internal_error")
				return
			}

			if !belowLimit {
				e.SendJSONError(w, http.StatusTooManyRequests, "too_many_attempts")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// checkRateLimit checks if the rate limit for the given key has been exceeded
func checkRateLimit(ctx context.Context, client *redis.Client, key string, expires time.Duration, maxCount int64) (bool, error) {
	pipe := client.Pipeline()

	// SET key 0 EX expires NX (only set if not exists)
	pipe.SetNX(ctx, key, 0, expires)

	// INCR key
	incrCmd := pipe.Incr(ctx, key)

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return false, fmt.Errorf("redis pipeline error: %w", err)
	}

	// Get the incremented value
	count, err := incrCmd.Result()
	if err != nil {
		return false, fmt.Errorf("failed to get incr result: %w", err)
	}

	return count <= maxCount, nil
}

// SimpleAuthCheck 创建一个权限验证中间件
// excludedPaths: 需要跳过验证的路径前缀列表
// authService: 用于验证 token 的服务
func SimpleAuthCheck(authService *services.AuthService, excludedPaths ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// 检查是否需要跳过验证
			if shouldSkip(path, excludedPaths) {
				next.ServeHTTP(w, r)
				return
			}

			// 尝试从 Cookie 或 Authorization header 中提取 token
			token := getTokenFromCookie(r, "token")
			if token == "" {
				token = extractBearerToken(r)
			}

			// 如果没有提供 token
			if token == "" {
				e.SendJSONError(w, 400, "bad_request", "no token provided")
				return
			}

			// 验证 token
			if !authService.IsValidToken(token) {
				e.SendJSONError(w, 401, "unauthorized", "invalid token")
				return
			}

			// token 有效，继续处理请求
			next.ServeHTTP(w, r)
		})
	}
}

// shouldSkip 检查给定路径是否应该跳过验证
func shouldSkip(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// extractBearerToken 从 Authorization header 中提取 Bearer token
func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// 检查是否以 "Bearer " 开头
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return ""
	}

	return strings.TrimPrefix(authHeader, bearerPrefix)
}

// getTokenFromCookie 从 Cookie 中获取指定名称的值
func getTokenFromCookie(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}
