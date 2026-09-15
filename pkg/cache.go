package pkg

import (
	"sync"
	"time"
)

// Cache provides a simple caching layer with TTL support
type Cache struct {
	mu       sync.RWMutex
	entries  map[string]*CacheEntry
	maxSize  int
	defaultTTL time.Duration
}

// CacheEntry represents a cached value with expiration
type CacheEntry struct {
	Value     string
	ExpiresAt time.Time
	Hits      uint64
}

// NewCache creates a new cache with specified max size
func NewCache(maxSize int, defaultTTL time.Duration) *Cache {
	if maxSize <= 0 {
		maxSize = 1000
	}
	if defaultTTL == 0 {
		defaultTTL = 5 * time.Minute
	}

	return &Cache{
		entries:    make(map[string]*CacheEntry),
		maxSize:    maxSize,
		defaultTTL: defaultTTL,
	}
}

// Set sets a value in the cache with default TTL
func (c *Cache) Set(key, value string) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL sets a value in the cache with custom TTL
func (c *Cache) SetWithTTL(key, value string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict if cache is full
	if len(c.entries) >= c.maxSize {
		c.evictOldest()
	}

	c.entries[key] = &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
		Hits:      0,
	}
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[key]
	if !ok {
		return "", false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		delete(c.entries, key)
		return "", false
	}

	entry.Hits++
	return entry.Value, true
}

// Delete removes a key from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// Clear clears all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*CacheEntry)
}

// evictOldest removes the oldest entry from the cache
func (c *Cache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for k, v := range c.entries {
		if oldestTime.IsZero() || v.ExpiresAt.Before(oldestTime) {
			oldestKey = k
			oldestTime = v.ExpiresAt
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
	}
}

// Size returns the number of entries in the cache
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// CleanupExpired removes all expired entries
func (c *Cache) CleanupExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	now := time.Now()

	for k, v := range c.entries {
		if now.After(v.ExpiresAt) {
			delete(c.entries, k)
			count++
		}
	}

	return count
}
