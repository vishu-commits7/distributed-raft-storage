package pkg

import (
	"encoding/json"
	"fmt"
	"time"
)

// RaftConfig holds configuration for the Raft engine
type RaftConfig struct {
	NodeID             string
	DataDir            string
	BindAddr           string
	BindPort           int
	Heartbeat          time.Duration
	ElectionTimeout    time.Duration
	SnapInterval       uint64
	Bootstrap          bool // True if this is the first node
}

// ClusterInfo holds information about the cluster
type ClusterInfo struct {
	Leader    string
	Term      uint64
	State     string
	CommitIdx uint64
	LastIdx   uint64
}

// NodeStats holds statistics about a node
type NodeStats struct {
	NodeID             string
	IsLeader           bool
	Leader             string
	Term               uint64
	LastLogIndex       uint64
	LastLogTerm        uint64
	CommitIndex        uint64
	LastApplied        uint64
	Applied            uint64
	FSMPending         uint64
	NumPeers           int
	LastContactLeader  time.Duration
}

// Ensure the RaftConfig methods return the fields for debugging
func (c *RaftConfig) String() string {
	b, _ := json.Marshal(c)
	return string(b)
}
