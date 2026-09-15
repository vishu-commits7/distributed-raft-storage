# Distributed Raft Storage Engine - Development Guide

## Overview

This guide covers the internal architecture and design decisions of the distributed Raft storage engine.

## Core Components

### 1. RaftEngine (`pkg/raft_engine.go`)

The main orchestration layer that wraps the Raft consensus algorithm.

**Key Responsibilities:**
- Initialize Raft cluster with BoltDB persistence
- Manage leader/follower state transitions
- Apply write operations (Set, Delete) through Raft log
- Serve read operations (Get, Scan) directly from state machine
- Handle peer management (join/leave cluster)

**Interface:**
```go
func (e *RaftEngine) Set(key, value string) error
func (e *RaftEngine) Delete(key string) error
func (e *RaftEngine) Get(key string) (string, bool)
func (e *RaftEngine) Scan(prefix string) map[string]string
func (e *RaftEngine) IsLeader() bool
func (e *RaftEngine) AddPeer(nodeID, addr string) error
func (e *RaftEngine) RemovePeer(nodeID string) error
```

### 2. StateMachine (`pkg/state_machine.go`)

Implements the `raft.FSM` interface for applying log entries to an in-memory store.

**Key Responsibilities:**
- Apply log entries (Set, Delete, Clear operations)
- Provide read access to the key-value store
- Support snapshots for fast recovery
- Thread-safe access with RWMutex

**Log Entry Format:**
```json
{
  "op": "set|delete|clear",
  "key": "string",
  "value": "string"
}
```

### 3. gRPC Server (`pkg/grpc_server.go`)

Provides client API for storage operations.

**Services:**
- `Get(key)` - Read from any node (eventual consistency)
- `Set(key, value)` - Write to leader only
- `Delete(key)` - Delete from leader only
- `Scan(prefix)` - Stream results with prefix matching

### 4. ClusterManager (`pkg/cluster_manager.go`)

Manages cluster membership and peer health.

**Features:**
- Track peer states (alive, suspected, dead)
- Add/remove peers dynamically
- Monitor last seen timestamps

### 5. MetricsCollector (`pkg/metrics.go`)

Collects operational metrics.

**Metrics:**
- Operation counts (Sets, Gets, Deletes)
- Latency averages
- Uptime

## Data Flow

### Write Path (Set)
```
Client Request
    ↓
Server checks: IsLeader?
    ↓ Yes
Serialize LogEntry (JSON)
    ↓
raft.Apply() → Raft Log
    ↓
Replication to Followers
    ↓
Quorum Committed
    ↓
StateMachine.Apply()
    ↓
Response to Client
```

### Read Path (Get)
```
Client Request
    ↓
StateMachine.Get(key)
    ↓
Return value (no Raft overhead)
```

## Persistence

**Storage Layout:**
```
/tmp/raft-node/
├── raft-log.db       # BoltDB store for Raft logs
├── raft-stable.db    # BoltDB store for stable state
└── snapshot-*.meta   # Snapshot metadata and data
```

**BoltDB Usage:**
- Log store: All Raft log entries
- Stable store: Current term, voted for candidate
- Snapshot store: Periodic snapshots of state machine

## Consensus Algorithm

### Raft Terms and Voting

1. **Election Timeout** (configurable, default 1s):
   - If follower doesn't hear from leader, starts election
   - Increments term, votes for itself
   - Requests votes from other peers

2. **Heartbeat** (configurable, default 100ms):
   - Leader sends empty AppendEntries RPC to followers
   - Maintains leadership and synchronizes state

3. **Log Replication**:
   - Leader receives write, appends to log
   - Sends to followers
   - Waits for majority confirmation
   - Commits and applies to state machine

## Configuration

**RaftConfig struct:**
```go
type RaftConfig struct {
    NodeID          string        // Unique node identifier
    DataDir         string        // Persistence directory
    BindAddr        string        // Network bind address
    BindPort        int           // Raft port
    Heartbeat       time.Duration // Leader heartbeat interval
    ElectionTimeout time.Duration // Election timeout
    SnapInterval    uint64        // Snapshot every N entries
    Bootstrap       bool          // First node in cluster
}
```

## Testing Strategy

### Unit Tests
- StateMachine operations (set, delete, scan)
- Snapshot persistence and restore
- Config validation

### Integration Tests
- Single-node Raft creation
- Set and get operations
- Multi-node replication (future)
- Leader election (future)
- Failure recovery (future)

## Performance Considerations

1. **Read Latency**: ~1-2ms (direct from state machine)
2. **Write Latency**: ~5-10ms (Raft replication overhead)
3. **Throughput**: ~10,000 ops/sec (single leader)
4. **Memory**: ~1MB per 10k entries (in-memory state machine)
5. **Disk**: ~100bytes per Raft log entry

## Future Enhancements

### Short Term
- [ ] Admin gRPC service for cluster management
- [ ] Prometheus metrics export
- [ ] Configuration hot-reload
- [ ] Better error handling and retries

### Medium Term
- [ ] REST API wrapper
- [ ] RocksDB backend option
- [ ] Compression for snapshots
- [ ] TTL support for keys

### Long Term
- [ ] Multi-region replication
- [ ] Encryption at rest and in transit
- [ ] Data compaction strategies
- [ ] Backup and restore CLI tools
- [ ] Web admin UI

## Debugging

### Enable Verbose Logging

Modify the logger in `NewRaftEngine`:
```go
logger := hclog.New(&hclog.LoggerOptions{
    Name:  "raft-engine",
    Level: hclog.Debug,  // Change to Debug
})
```

### Inspect State Machine

Add debug commands:
```go
func (e *RaftEngine) DumpState() {
    data := e.FSM.Scan("")
    for k, v := range data {
        log.Printf("%s: %s", k, v)
    }
}
```

## References

- [Raft Consensus Algorithm](https://raft.github.io/)
- [HashiCorp Raft Implementation](https://github.com/hashicorp/raft)
- [BoltDB Documentation](https://github.com/boltdb/bolt)
- [gRPC Go Tutorial](https://grpc.io/docs/languages/go/)
