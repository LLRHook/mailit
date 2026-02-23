package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/mailit-dev/mailit/internal/pkg"
)

type RateLimitConfig struct {
	Enabled    bool
	DefaultRPS int
	SendRPS    int
	BatchRPS   int
	Window     time.Duration
}

// RateLimit creates a Redis-backed rate limiter middleware.
func RateLimit(rdb *redis.Client, cfg RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			auth := GetAuth(r.Context())
			if auth == nil {
				next.ServeHTTP(w, r)
				return
			}

			limit := cfg.DefaultRPS
			key := fmt.Sprintf("ratelimit:%s:default", auth.TeamID.String())

			window := cfg.Window
			if window == 0 {
				window = time.Second
			}

			// Use a sliding window counter in Redis
			now := time.Now()
			windowKey := fmt.Sprintf("%s:%d", key, now.Unix())

			pipe := rdb.Pipeline()
			incr := pipe.Incr(r.Context(), windowKey)
			pipe.Expire(r.Context(), windowKey, window*2)
			_, err := pipe.Exec(r.Context())
			if err != nil {
				// If Redis is down, allow the request
				next.ServeHTTP(w, r)
				return
			}

			count := incr.Val()

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(max(0, limit-int(count))))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(now.Add(window).Unix(), 10))

			if int(count) > limit {
				w.Header().Set("Retry-After", strconv.Itoa(int(window.Seconds())))
				pkg.Error(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SendRateLimit applies a stricter rate limit for email sending endpoints.
func SendRateLimit(rdb *redis.Client, cfg RateLimitConfig) func(http.Handler) http.Handler {
	sendCfg := cfg
	sendCfg.DefaultRPS = cfg.SendRPS
	return RateLimit(rdb, sendCfg)
}

// BatchRateLimit applies the strictest rate limit for batch operations.
func BatchRateLimit(rdb *redis.Client, cfg RateLimitConfig) func(http.Handler) http.Handler {
	batchCfg := cfg
	batchCfg.DefaultRPS = cfg.BatchRPS
	return RateLimit(rdb, batchCfg)
}
