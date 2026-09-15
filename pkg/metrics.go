package pkg

import (
	"sync"
	"time"
)

// MetricsCollector collects statistics about the Raft engine
type MetricsCollector struct {
	mu sync.RWMutex

	// Operations
	Sets   uint64
	Gets   uint64
	Deletes uint64

	// Performance
	AvgWriteLatency uint64 // microseconds
	AvgReadLatency  uint64 // microseconds

	// Raft
	LastLogIndex uint64
	LastLogTerm  uint64
	CommitIndex  uint64

	// Timestamps
	StartTime time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		StartTime: time.Now(),
	}
}

// RecordSet records a set operation
func (mc *MetricsCollector) RecordSet(latency time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.Sets++
	mc.AvgWriteLatency = (mc.AvgWriteLatency + uint64(latency.Microseconds())) / 2
}

// RecordGet records a get operation
func (mc *MetricsCollector) RecordGet(latency time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.Gets++
	mc.AvgReadLatency = (mc.AvgReadLatency + uint64(latency.Microseconds())) / 2
}

// RecordDelete records a delete operation
func (mc *MetricsCollector) RecordDelete(latency time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.Deletes++
}

// GetMetrics returns a copy of current metrics
func (mc *MetricsCollector) GetMetrics() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return map[string]interface{}{
		"sets":                mc.Sets,
		"gets":                mc.Gets,
		"deletes":             mc.Deletes,
		"avg_write_latency_us": mc.AvgWriteLatency,
		"avg_read_latency_us":  mc.AvgReadLatency,
		"uptime_seconds":       int64(time.Since(mc.StartTime).Seconds()),
	}
}
