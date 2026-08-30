package http

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// SecurityHeadersMiddleware attaches defensive HTTP headers to all responses.
func SecurityHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware configures cross-origin request handling.
func CORSMiddleware(allowAll bool) func(http.Handler) http.Handler {
	allowedOrigins := []string{"*"}
	if !allowAll {
		allowedOrigins = []string{"http://localhost:*", "http://127.0.0.1:*"}
	}

	return cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Range"},
		ExposedHeaders:   []string{"Link", "Content-Length", "Content-Range", "Accept-Ranges", "ETag"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}

// SlogRequestLogger logs incoming HTTP requests using standard slog.
func SlogRequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				latency := time.Since(start)
				reqID := middleware.GetReqID(r.Context())

				level := slog.LevelInfo
				if ww.Status() >= 500 {
					level = slog.LevelError
				} else if ww.Status() >= 400 {
					level = slog.LevelWarn
				}

				logger.Log(r.Context(), level, "HTTP request",
					"method", r.Method,
					"path", r.URL.Path,
					"status", ww.Status(),
					"duration_ms", latency.Milliseconds(),
					"bytes_written", ww.BytesWritten(),
					"remote_ip", r.RemoteAddr,
					"request_id", reqID,
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

// IPRateLimiter provides token-bucket rate limiting based on client IP.
type IPRateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewIPRateLimiter(limit int, windowSec int) *IPRateLimiter {
	return &IPRateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   time.Duration(windowSec) * time.Second,
	}
}

func (rl *IPRateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				ip = forwarded
			}

			rl.mu.Lock()
			now := time.Now()
			cutoff := now.Add(-rl.window)

			// Purge expired entries
			var valid []time.Time
			for _, t := range rl.requests[ip] {
				if t.After(cutoff) {
					valid = append(valid, t)
				}
			}

			if len(valid) >= rl.limit {
				rl.requests[ip] = valid
				rl.mu.Unlock()
				w.Header().Set("Retry-After", "60")
				JSON(w, http.StatusTooManyRequests, ErrorResponse{
					Error: ErrorPayload{
						Code:    "RATE_LIMIT_EXCEEDED",
						Message: "Too many requests. Please wait before retrying.",
					},
				})
				return
			}

			valid = append(valid, now)
			rl.requests[ip] = valid
			rl.mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
