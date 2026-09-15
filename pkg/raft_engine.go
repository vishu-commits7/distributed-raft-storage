package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	ratboltdb "github.com/hashicorp/raft-boltdb"
)

// RaftEngine encapsulates the Raft cluster and state machine
type RaftEngine struct {
	raft      *raft.Raft
	FSM       *StateMachine
	logStore  raft.LogStore
	stableStore raft.StableStore
	snapStore raft.SnapshotStore
	config    *RaftConfig
	logger    hclog.Logger
}

// RaftConfig holds configuration for the Raft engine
type RaftConfig struct {
	NodeID      string
	DataDir     string
	BindAddr    string
	BindPort    int
	Heartbeat   time.Duration
	Election    time.Duration
	SnapInterval uint64
	Bootstrap   bool // True if this is the first node
}

// NewRaftEngine creates a new Raft engine
func NewRaftEngine(cfg *RaftConfig) (*RaftEngine, error) {
	if cfg.Heartbeat == 0 {
		cfg.Heartbeat = 100 * time.Millisecond
	}
	if cfg.ElectionTimeout == 0 {
		cfg.ElectionTimeout = 1 * time.Second
	}
	if cfg.SnapInterval == 0 {
		cfg.SnapInterval = 8192
	}

	// Create logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "raft-engine",
		Level: hclog.Info,
	})

	// Create data directory
	if err := os.MkdirAll(cfg.DataDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Create log store
	logStore, err := ratboltdb.NewBoltStore(filepath.Join(cfg.DataDir, "raft-log.db"))
	if err != nil {
		return nil, fmt.Errorf("failed to create log store: %w", err)
	}

	// Create stable store
	stableStore, err := ratboltdb.NewBoltStore(filepath.Join(cfg.DataDir, "raft-stable.db"))
	if err != nil {
		return nil, fmt.Errorf("failed to create stable store: %w", err)
	}

	// Create snapshot store
	snapStore, err := raft.NewFileSnapshotStore(cfg.DataDir, 3, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot store: %w", err)
	}

	// Create state machine
	fsm := NewStateMachine()

	// Create Raft configuration
	ratCfg := raft.DefaultConfig()
	ratCfg.ProtocolVersion = raft.ProtocolVersionMax
	ratCfg.HeartbeatTimeout = cfg.Heartbeat
	ratCfg.ElectionTimeout = cfg.ElectionTimeout
	ratCfg.SnapshotInterval = cfg.SnapInterval
	ratCfg.Logger = logger

	// Bind transport
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", cfg.BindAddr, cfg.BindPort))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address: %w", err)
	}

	transport, err := raft.NewTCPTransport(addr.String(), addr, 3, 10*time.Second, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	// Create Raft instance
	raftNode, err := raft.NewRaft(ratCfg, fsm, logStore, stableStore, snapStore, transport)
	if err != nil {
		return nil, fmt.Errorf("failed to create raft: %w", err)
	}

	// Bootstrap if needed
	if cfg.Bootstrap {
		cfg := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raft.ServerID(cfg.NodeID),
					Address: raft.ServerAddress(addr.String()),
				},
			},
		}
		future := raftNode.BootstrapCluster(cfg)
		if err := future.Error(); err != nil {
			return nil, fmt.Errorf("failed to bootstrap cluster: %w", err)
		}
	}

	return &RaftEngine{
		raft:        raftNode,
		FSM:         fsm,
		logStore:    logStore,
		stableStore: stableStore,
		snapStore:   snapStore,
		config:      cfg,
		logger:      logger,
	}, nil
}

// Set applies a set command to the Raft cluster
func (e *RaftEngine) Set(key, value string) error {
	if !e.IsLeader() {
		return fmt.Errorf("not a leader")
	}

	entry := LogEntry{
		Op:    "set",
		Key:   key,
		Value: value,
	}

	b, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	future := e.raft.Apply(b, 5*time.Second)
	return future.Error()
}

// Delete applies a delete command to the Raft cluster
func (e *RaftEngine) Delete(key string) error {
	if !e.IsLeader() {
		return fmt.Errorf("not a leader")
	}

	entry := LogEntry{
		Op:  "delete",
		Key: key,
	}

	b, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	future := e.raft.Apply(b, 5*time.Second)
	return future.Error()
}

// Get retrieves a value from the state machine (read-only)
func (e *RaftEngine) Get(key string) (string, bool) {
	return e.FSM.Get(key)
}

// Scan returns all keys with the given prefix
func (e *RaftEngine) Scan(prefix string) map[string]string {
	return e.FSM.Scan(prefix)
}

// IsLeader returns true if this node is the leader
func (e *RaftEngine) IsLeader() bool {
	return e.raft.State() == raft.Leader
}

// GetLeader returns the current leader
func (e *RaftEngine) GetLeader() string {
	return string(e.raft.Leader())
}

// GetState returns the current Raft state
func (e *RaftEngine) GetState() raft.RaftState {
	return e.raft.State()
}

// AddPeer adds a new peer to the cluster
func (e *RaftEngine) AddPeer(nodeID, addr string) error {
	if !e.IsLeader() {
		return fmt.Errorf("not a leader")
	}

	future := e.raft.AddVoter(
		raft.ServerID(nodeID),
		raft.ServerAddress(addr),
		0,
		5*time.Second,
	)
	return future.Error()
}

// RemovePeer removes a peer from the cluster
func (e *RaftEngine) RemovePeer(nodeID string) error {
	if !e.IsLeader() {
		return fmt.Errorf("not a leader")
	}

	future := e.raft.RemoveServer(
		raft.ServerID(nodeID),
		0,
		5*time.Second,
	)
	return future.Error()
}

// Close closes the Raft engine
func (e *RaftEngine) Close() error {
	future := e.raft.Shutdown()
	if err := future.Error(); err != nil {
		return err
	}

	if err := e.logStore.Close(); err != nil {
		return err
	}

	if err := e.stableStore.Close(); err != nil {
		return err
	}

	return nil
}

// GetStats returns Raft statistics
func (e *RaftEngine) GetStats() map[string]string {
	return e.raft.Stats()
}
