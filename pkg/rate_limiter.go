package pkg

import (
	"fmt"
	"regexp"
	"sync"
)

// RateLimiter provides rate limiting for operations
type RateLimiter struct {
	mu       sync.RWMutex
	limits   map[string]int // client -> requests per second

ctomers map[string]*ClientRateStatus
}

// ClientRateStatus tracks rate limit status for a client
type ClientRateStatus struct {
	ClientID       string
	RequestCount   int
	LastResetTime  int64
	IsThrottled    bool
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limits:   make(map[string]int),
		customers: make(map[string]*ClientRateStatus),
	}
}

// SetLimit sets the rate limit for a client
func (rl *RateLimiter) SetLimit(clientID string, requestsPerSec int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if requestsPerSec <= 0 {
		delete(rl.limits, clientID)
	} else {
		rl.limits[clientID] = requestsPerSec
	}
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limit, hasLimit := rl.limits[clientID]
	if !hasLimit {
		return true // No limit set
	}

	status, ok := rl.customers[clientID]
	if !ok {
		status = &ClientRateStatus{
			ClientID:      clientID,
			RequestCount:  0,
			LastResetTime: getTimestamp(),
			IsThrottled:   false,
		}
		rl.customers[clientID] = status
	}

	now := getTimestamp()

	// Reset counter if a second has passed
	if now-status.LastResetTime >= 1000 {
		status.RequestCount = 0
		status.LastResetTime = now
		status.IsThrottled = false
	}

	// Check if limit exceeded
	if status.RequestCount >= limit {
		status.IsThrottled = true
		return false
	}

	status.RequestCount++
	return true
}

// GetStatus returns the rate limit status for a client
func (rl *RateLimiter) GetStatus(clientID string) *ClientRateStatus {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	if status, ok := rl.customers[clientID]; ok {
		copy := *status
		return &copy
	}
	return nil
}

// RemoveLimit removes the rate limit for a client
func (rl *RateLimiter) RemoveLimit(clientID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.limits, clientID)
}

// ValidateClientID checks if a client ID matches expected pattern
func (rl *RateLimiter) ValidateClientID(clientID string) bool {
	pattern := `^[a-zA-Z0-9_-]+$`
	match, _ := regexp.MatchString(pattern, clientID)
	return match
}

// getTimestamp returns current time in milliseconds
func getTimestamp() int64 {
	return int64(0) // Placeholder, implement with time.Now()
}
