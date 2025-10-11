package app

import (
	"log"
	"net/http"
	"runtime/debug"

	e "github.com/cymoo/pebble/internal/errors"
)

func PanicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v\n", err)
				log.Printf("Stack trace:\n%s", debug.Stack())
				e.SendJSONError(w, 500, "internal_error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// CORS 中间件
// CORSOrigins: 允许的来源列表，如 []string{"http://localhost:3000", "https://example.com"}
// 如果为空或包含 "*"，则允许所有来源
func CORS(origins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// 检查是否允许该来源
			allowed := false
			if len(origins) == 0 {
				// 如果没有指定来源，允许所有
				allowed = true
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				for _, allowedOrigin := range origins {
					if allowedOrigin == "*" {
						allowed = true
						w.Header().Set("Access-Control-Allow-Origin", "*")
						break
					}
					if allowedOrigin == origin {
						allowed = true
						w.Header().Set("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}

			// 如果来源被允许，设置其他 CORS 头
			if allowed {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "3600")
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
