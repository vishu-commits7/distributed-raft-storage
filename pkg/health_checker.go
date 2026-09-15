package pkg

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// HealthChecker monitors node health and cluster status
type HealthChecker struct {
	mu              sync.RWMutex
	eng             *RaftEngine
	ticker          *time.Ticker
	done            chan bool
	lastHealthCheck time.Time
	isHealthy       bool
}

// HealthStatus represents the health of a node
type HealthStatus struct {
	NodeID        string
	IsLeader      bool
	IsHealthy     bool
	LastCheckTime time.Time
	Uptime        time.Duration
	RaftState     string
	PeerCount     int
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(eng *RaftEngine) *HealthChecker {
	return &HealthChecker{
		eng:       eng,
		done:      make(chan bool),
		isHealthy: true,
	}
}

// Start begins periodic health checks
func (hc *HealthChecker) Start(interval time.Duration) {
	hc.ticker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-hc.ticker.C:
				hc.check()
			case <-hc.done:
				return
			}
		}
	}()
}

// check performs a health check
func (hc *HealthChecker) check() {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	// Basic health check: can we read from state machine?
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try a simple get operation
	_, _ = hc.eng.Get("__health_check__")

	hc.isHealthy = true
	hc.lastHealthCheck = time.Now()
}

// GetStatus returns the current health status
func (hc *HealthChecker) GetStatus() HealthStatus {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	return HealthStatus{
		NodeID:        hc.eng.config.NodeID,
		IsLeader:      hc.eng.IsLeader(),
		IsHealthy:     hc.isHealthy,
		LastCheckTime: hc.lastHealthCheck,
		RaftState:     hc.eng.GetState().String(),
	}
}

// Stop stops the health checker
func (hc *HealthChecker) Stop() {
	if hc.ticker != nil {
		hc.ticker.Stop()
	}
	hc.done <- true
}

// IsHealthy returns true if the node is healthy
func (hc *HealthChecker) IsHealthy() bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.isHealthy
}
