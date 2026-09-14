package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"hzycoder.com/lion/pkg/response"
)

const limiterEntryTTL = 10 * time.Minute

type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter 基于 IP 的内存令牌桶限速器
type RateLimiter struct {
	mu       sync.Mutex
	entries  map[string]*rateLimiterEntry
	interval time.Duration
	burst    int
}

// NewRateLimiter requestsPerMinute 为单 IP 每分钟允许的最大请求数
func NewRateLimiter(requestsPerMinute, burst int) *RateLimiter {
	if burst <= 0 {
		burst = requestsPerMinute
	}
	return &RateLimiter{
		entries:  make(map[string]*rateLimiterEntry),
		interval: time.Minute / time.Duration(requestsPerMinute),
		burst:    burst,
	}
}

// Limit 返回 Gin 中间件；receiver 为 nil 时不做任何限制
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	if rl == nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()
		now := time.Now()
		for key, entry := range rl.entries {
			if now.Sub(entry.lastSeen) > limiterEntryTTL {
				delete(rl.entries, key)
			}
		}
		entry, ok := rl.entries[ip]
		if !ok {
			entry = &rateLimiterEntry{limiter: rate.NewLimiter(rate.Every(rl.interval), rl.burst)}
			rl.entries[ip] = entry
		}
		entry.lastSeen = now
		rl.mu.Unlock()

		if !entry.limiter.Allow() {
			response.AbortWithCode(c, http.StatusTooManyRequests, response.CodeTooManyRequests)
			return
		}
		c.Next()
	}
}
