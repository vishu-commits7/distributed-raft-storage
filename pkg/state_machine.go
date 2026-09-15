package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/raft"
)

// LogEntry represents a single state machine command
type LogEntry struct {
	Op    string `json:"op"`    // "set", "delete", "clear"
	Key   string `json:"key"`
	Value string `json:"value"`
}

// StateMachine implements raft.FSM interface
type StateMachine struct {
	mu    sync.RWMutex
	store map[string]string
}

// NewStateMachine creates a new state machine
func NewStateMachine() *StateMachine {
	return &StateMachine{
		store: make(map[string]string),
	}
}

// Apply applies a log entry to the state machine
func (sm *StateMachine) Apply(log *raft.Log) interface{} {
	var entry LogEntry
	if err := json.Unmarshal(log.Data, &entry); err != nil {
		return fmt.Errorf("failed to unmarshal log entry: %w", err)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	switch entry.Op {
	case "set":
		sm.store[entry.Key] = entry.Value
		return nil
	case "delete":
		delete(sm.store, entry.Key)
		return nil
	case "clear":
		sm.store = make(map[string]string)
		return nil
	default:
		return fmt.Errorf("unknown operation: %s", entry.Op)
	}
}

// Snapshot returns a snapshot of the state machine
func (sm *StateMachine) Snapshot() (raft.FSMSnapshot, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Create a copy of the store
	storeCopy := make(map[string]string)
	for k, v := range sm.store {
		storeCopy[k] = v
	}

	return &StateMachineSnapshot{store: storeCopy}, nil
}

// Restore restores the state machine from a snapshot
func (sm *StateMachine) Restore(snapshot io.ReadCloser) error {
	defer snapshot.Close()

	var store map[string]string
	if err := json.NewDecoder(snapshot).Decode(&store); err != nil {
		return fmt.Errorf("failed to decode snapshot: %w", err)
	}

	sm.mu.Lock()
	sm.store = store
	sm.mu.Unlock()

	return nil
}

// Get retrieves a value from the state machine
func (sm *StateMachine) Get(key string) (string, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	val, ok := sm.store[key]
	return val, ok
}

// Scan returns all keys with the given prefix
func (sm *StateMachine) Scan(prefix string) map[string]string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make(map[string]string)
	for k, v := range sm.store {
		if len(prefix) == 0 || len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			result[k] = v
		}
	}
	return result
}

// StateMachineSnapshot represents a snapshot of the state machine
type StateMachineSnapshot struct {
	store map[string]string
}

// Persist writes the snapshot to a sink
func (s *StateMachineSnapshot) Persist(sink raft.SnapshotSink) error {
	defer sink.Close()
	b, err := json.Marshal(s.store)
	if err != nil {
		return err
	}
	_, err = sink.Write(b)
	return err
}

// Release releases the snapshot
func (s *StateMachineSnapshot) Release() {
	// No-op for in-memory snapshot
}
