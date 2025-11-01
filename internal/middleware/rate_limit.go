package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter manages rate limiting for requests
type RateLimiter struct {
	requests map[string]*clientRequests
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

// clientRequests tracks requests for a specific client IP
type clientRequests struct {
	count     int
	firstSeen time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*clientRequests),
		limit:    limit,
		window:   window,
	}

	// Start cleanup goroutine
	go rl.cleanupExpiredEntries()

	return rl
}

// cleanupExpiredEntries removes expired entries from the rate limiter
func (rl *RateLimiter) cleanupExpiredEntries() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, req := range rl.requests {
			if now.Sub(req.firstSeen) > rl.window {
				delete(rl.requests, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware returns a Gin middleware handler for rate limiting
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP
		clientIP := c.ClientIP()

		// Check rate limit
		if !rl.allowRequest(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
				"code":  "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// allowRequest checks if a request from the given IP should be allowed
func (rl *RateLimiter) allowRequest(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Get or create client request record
	req, exists := rl.requests[clientIP]
	if !exists {
		rl.requests[clientIP] = &clientRequests{
			count:     1,
			firstSeen: now,
		}
		return true
	}

	// Check if the time window has passed
	if now.Sub(req.firstSeen) > rl.window {
		// Reset the counter for a new window
		req.count = 1
		req.firstSeen = now
		return true
	}

	// Check if limit is exceeded
	if req.count >= rl.limit {
		return false
	}

	// Increment counter and allow request
	req.count++
	return true
}

// RateLimit creates a rate limiting middleware with the specified limits
func RateLimit(requests int, duration time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(requests, duration)
	return limiter.Middleware()
}
