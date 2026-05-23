package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*tokenBucket
	rate     float64 // tokens per second
	capacity float64 // max tokens
}

type tokenBucket struct {
	tokens   float64
	lastTime time.Time
}

func NewRateLimiter(rate, capacity float64) *RateLimiter {
	return &RateLimiter{
		buckets:  make(map[string]*tokenBucket),
		rate:     rate,
		capacity: capacity,
	}
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[key]
	now := time.Now()

	if !exists {
		rl.buckets[key] = &tokenBucket{tokens: rl.capacity - 1, lastTime: now}
		return true
	}

	elapsed := now.Sub(bucket.lastTime).Seconds()
	bucket.tokens += elapsed * rl.rate
	if bucket.tokens > rl.capacity {
		bucket.tokens = rl.capacity
	}
	bucket.lastTime = now

	if bucket.tokens >= 1 {
		bucket.tokens--
		return true
	}
	return false
}

func RateLimit(rate, capacity float64) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, capacity)

	// Cleanup old buckets periodically
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			limiter.mu.Lock()
			cutoff := time.Now().Add(-30 * time.Minute)
			for k, v := range limiter.buckets {
				if v.lastTime.Before(cutoff) {
					delete(limiter.buckets, k)
				}
			}
			limiter.mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		key := c.ClientIP()
		if !limiter.allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    42901,
				"message": "rate limit exceeded, try again later",
				"data":    nil,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
