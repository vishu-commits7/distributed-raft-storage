package pkg

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
)

// Snapshot represents a consistent view of the state machine at a point in time
type Snapshot struct {
	ID        string            // Unique snapshot ID
	Timestamp int64             // Unix timestamp
	LogIndex  uint64            // Raft log index
	LogTerm   uint64            // Raft log term
	Data      map[string]string // Key-value pairs
	Checksum  string            // SHA256 hash for verification
}

// SnapshotManager handles creation and restoration of snapshots
type SnapshotManager struct {
	mu        sync.RWMutex
	snapshots map[string]*Snapshot
	eng       *RaftEngine
}

// NewSnapshotManager creates a new snapshot manager
func NewSnapshotManager(eng *RaftEngine) *SnapshotManager {
	return &SnapshotManager{
		snapshots: make(map[string]*Snapshot),
		eng:       eng,
	}
}

// Create creates a new snapshot of the current state
func (sm *SnapshotManager) Create(id string) (*Snapshot, error) {
	data := sm.eng.Scan("") // Get all data

	// Calculate checksum
	hash := sha256.New()
	for k, v := range data {
		hash.Write([]byte(k))
		hash.Write([]byte(v))
	}
	checksum := hex.EncodeToString(hash.Sum(nil))

	snapshot := &Snapshot{
		ID:       id,
		Timestamp: sm.eng.raft.LastContact().Unix(),
		Data:     data,
		Checksum: checksum,
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.snapshots[id] = snapshot
	return snapshot, nil
}

// Get retrieves a snapshot by ID
func (sm *SnapshotManager) Get(id string) *Snapshot {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if snap, ok := sm.snapshots[id]; ok {
		// Return a copy
		copy := *snap
		copy.Data = make(map[string]string)
		for k, v := range snap.Data {
			copy.Data[k] = v
		}
		return &copy
	}
	return nil
}

// List returns all snapshot IDs
func (sm *SnapshotManager) List() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	ids := make([]string, 0, len(sm.snapshots))
	for id := range sm.snapshots {
		ids = append(ids, id)
	}
	return ids
}

// Verify verifies the integrity of a snapshot
func (sm *SnapshotManager) Verify(id string) (bool, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	snapshot, ok := sm.snapshots[id]
	if !ok {
		return false, fmt.Errorf("snapshot not found: %s", id)
	}

	// Recalculate checksum
	hash := sha256.New()
	for k, v := range snapshot.Data {
		hash.Write([]byte(k))
		hash.Write([]byte(v))
	}
	calculated := hex.EncodeToString(hash.Sum(nil))

	return calculated == snapshot.Checksum, nil
}

// Delete deletes a snapshot
func (sm *SnapshotManager) Delete(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, ok := sm.snapshots[id]; !ok {
		return fmt.Errorf("snapshot not found: %s", id)
	}

	delete(sm.snapshots, id)
	return nil
}

// Size returns the number of snapshots
func (sm *SnapshotManager) Size() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.snapshots)
}
