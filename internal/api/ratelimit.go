package api

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a simple in-memory rate limiter using token bucket algorithm
type RateLimiter struct {
	requestsPerMinute int
	buckets           sync.Map // map[string]*tokenBucket
	cleanupInterval   time.Duration
	stopCleanup       chan struct{}
	mu                sync.Mutex
}

// tokenBucket represents a token bucket for rate limiting
type tokenBucket struct {
	tokens     int
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		requestsPerMinute: requestsPerMinute,
		cleanupInterval:   5 * time.Minute, // Cleanup stale entries every 5 minutes
		stopCleanup:       make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.startCleanup()

	return rl
}

// Allow checks if a request should be allowed based on rate limiting
// Returns true if allowed, false if rate limited
func (rl *RateLimiter) Allow(identifier string) bool {
	now := time.Now()
	tokensPerSecond := float64(rl.requestsPerMinute) / 60.0

	// Get or create bucket for this identifier
	value, _ := rl.buckets.LoadOrStore(identifier, &tokenBucket{
		tokens:     rl.requestsPerMinute,
		lastRefill: now,
	})

	bucket := value.(*tokenBucket)
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Calculate time elapsed since last refill
	elapsed := now.Sub(bucket.lastRefill).Seconds()

	// Refill tokens based on elapsed time
	tokensToAdd := int(elapsed * tokensPerSecond)
	if tokensToAdd > 0 {
		bucket.tokens += tokensToAdd
		// Cap tokens at requestsPerMinute
		if bucket.tokens > rl.requestsPerMinute {
			bucket.tokens = rl.requestsPerMinute
		}
		bucket.lastRefill = now
	}

	// Check if we have tokens available
	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

// startCleanup periodically removes stale buckets to prevent memory leaks
func (rl *RateLimiter) startCleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopCleanup:
			return
		}
	}
}

// cleanup removes buckets that haven't been used in the last hour
func (rl *RateLimiter) cleanup() {
	cutoff := time.Now().Add(-1 * time.Hour)
	rl.buckets.Range(func(key, value interface{}) bool {
		bucket := value.(*tokenBucket)
		bucket.mu.Lock()
		lastRefill := bucket.lastRefill
		bucket.mu.Unlock()

		if lastRefill.Before(cutoff) {
			rl.buckets.Delete(key)
		}
		return true
	})
}

// Stop stops the cleanup goroutine
func (rl *RateLimiter) Stop() {
	close(rl.stopCleanup)
}

// GetClientIP extracts the client IP address from the request (exported for use in main handler)
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (used by proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := []string{}
		for _, ip := range splitComma(xff) {
			ips = append(ips, trimSpace(ip))
		}
		if len(ips) > 0 {
			return ips[0]
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return trimSpace(xri)
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := indexLast(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// Helper functions for string manipulation
func splitComma(s string) []string {
	result := []string{}
	current := ""
	for _, char := range s {
		if char == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func indexLast(s string, substr string) int {
	subLen := len(substr)
	if subLen == 0 {
		return len(s)
	}
	for i := len(s) - subLen; i >= 0; i-- {
		match := true
		for j := 0; j < subLen; j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

