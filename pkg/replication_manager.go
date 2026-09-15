package pkg

import (
	"fmt"
	"sync"
	"time"
)

// ReplicationManager handles data replication across cluster nodes
type ReplicationManager struct {
	mu            sync.RWMutex
	eng           *RaftEngine
	replicationCh chan ReplicationEvent
	peerStatus    map[string]*PeerReplicationStatus
}

// ReplicationEvent represents a replication event
type ReplicationEvent struct {
	Timestamp  time.Time
	SourceNode string
	TargetNode string
	EventType  string // "sync", "resync", "lag"
	DataCount  int
	Error      string
}

// PeerReplicationStatus tracks replication status for a peer
type PeerReplicationStatus struct {
	PeerID           string
	IsInSync         bool
	LastSyncTime     time.Time
	SyncLag          int64 // in milliseconds
	FailedSyncCount  int
	SuccessSyncCount int
}

// NewReplicationManager creates a new replication manager
func NewReplicationManager(eng *RaftEngine) *ReplicationManager {
	return &ReplicationManager{
		eng:           eng,
		replicationCh: make(chan ReplicationEvent, 100),
		peerStatus:    make(map[string]*PeerReplicationStatus),
	}
}

// SyncPeer syncs data with a specific peer
func (rm *ReplicationManager) SyncPeer(peerID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if !rm.eng.IsLeader() {
		return fmt.Errorf("only leader can initiate sync")
	}

	status, ok := rm.peerStatus[peerID]
	if !ok {
		status = &PeerReplicationStatus{
			PeerID:       peerID,
			IsInSync:     false,
			LastSyncTime: time.Now(),
		}
		rm.peerStatus[peerID] = status
	}

	// Simulate sync
	event := ReplicationEvent{
		Timestamp:  time.Now(),
		SourceNode: rm.eng.config.NodeID,
		TargetNode: peerID,
		EventType:  "sync",
		DataCount:  len(rm.eng.Scan("")),
	}

	rm.replicationCh <- event

	status.IsInSync = true
	status.LastSyncTime = time.Now()
	status.SuccessSyncCount++

	return nil
}

// GetPeerStatus returns replication status for a peer
func (rm *ReplicationManager) GetPeerStatus(peerID string) *PeerReplicationStatus {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if status, ok := rm.peerStatus[peerID]; ok {
		copy := *status
		return &copy
	}
	return nil
}

// GetAllPeerStatus returns status for all peers
func (rm *ReplicationManager) GetAllPeerStatus() map[string]*PeerReplicationStatus {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	copy := make(map[string]*PeerReplicationStatus)
	for k, v := range rm.peerStatus {
		copy[k] = v
	}
	return copy
}

// GetReplicationEvents returns pending replication events
func (rm *ReplicationManager) GetReplicationEvents(maxCount int) []ReplicationEvent {
	var events []ReplicationEvent
	for i := 0; i < maxCount; i++ {
		select {
		case event := <-rm.replicationCh:
			events = append(events, event)
		default:
			return events
		}
	}
	return events
}
