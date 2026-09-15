package pkg

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// AuditLog records all operations for accountability and debugging
type AuditLog struct {
	mu      sync.RWMutex
	entries []AuditEntry
	maxSize int
}

// AuditEntry represents a single audit log entry
type AuditEntry struct {
	Timestamp time.Time
	NodeID    string
	Operation string // "set", "get", "delete", "scan"
	Key       string
	OldValue  string
	NewValue  string
	Success   bool
	Error     string
	ActorIP   string
}

// NewAuditLog creates a new audit log with specified max size
func NewAuditLog(maxSize int) *AuditLog {
	if maxSize <= 0 {
		maxSize = 10000 // default to 10k entries
	}
	return &AuditLog{
		entries: make([]AuditEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Log records an operation in the audit log
func (al *AuditLog) Log(entry AuditEntry) {
	al.mu.Lock()
	defer al.mu.Unlock()

	entry.Timestamp = time.Now()
	al.entries = append(al.entries, entry)

	// Keep only the last maxSize entries (circular buffer)
	if len(al.entries) > al.maxSize {
		al.entries = al.entries[len(al.entries)-al.maxSize:]
	}
}

// GetEntries returns all audit log entries
func (al *AuditLog) GetEntries() []AuditEntry {
	al.mu.RLock()
	defer al.mu.RUnlock()

	// Return a copy
	entries := make([]AuditEntry, len(al.entries))
	copy(entries, al.entries)
	return entries
}

// GetEntriesByOperation returns entries for a specific operation
func (al *AuditLog) GetEntriesByOperation(op string) []AuditEntry {
	al.mu.RLock()
	defer al.mu.RUnlock()

	var result []AuditEntry
	for _, e := range al.entries {
		if e.Operation == op {
			result = append(result, e)
		}
	}
	return result
}

// GetEntriesByKey returns entries for a specific key
func (al *AuditLog) GetEntriesByKey(key string) []AuditEntry {
	al.mu.RLock()
	defer al.mu.RUnlock()

	var result []AuditEntry
	for _, e := range al.entries {
		if e.Key == key {
			result = append(result, e)
		}
	}
	return result
}

// GetEntriesSince returns entries after a given time
func (al *AuditLog) GetEntriesSince(since time.Time) []AuditEntry {
	al.mu.RLock()
	defer al.mu.RUnlock()

	var result []AuditEntry
	for _, e := range al.entries {
		if e.Timestamp.After(since) {
			result = append(result, e)
		}
	}
	return result
}

// Export exports the audit log to JSON
func (al *AuditLog) Export() (string, error) {
	al.mu.RLock()
	defer al.mu.RUnlock()

	b, err := json.MarshalIndent(al.entries, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal audit log: %w", err)
	}
	return string(b), nil
}

// Clear clears all audit log entries
func (al *AuditLog) Clear() {
	al.mu.Lock()
	defer al.mu.Unlock()
	al.entries = make([]AuditEntry, 0, al.maxSize)
}

// Size returns the number of entries in the audit log
func (al *AuditLog) Size() int {
	al.mu.RLock()
	defer al.mu.RUnlock()
	return len(al.entries)
}
