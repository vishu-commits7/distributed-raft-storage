package pkg

import (
	"fmt"
	"sync"
	"time"
)

// Peer represents a cluster peer
type Peer struct {
	ID      string
	Addr    string
	Status  string // "alive", "suspected", "dead"
	LastSeen time.Time
}

// ClusterManager handles cluster membership
type ClusterManager struct {
	mu    sync.RWMutex
	peers map[string]*Peer
	eng   *RaftEngine
}

// NewClusterManager creates a new cluster manager
func NewClusterManager(engine *RaftEngine) *ClusterManager {
	return &ClusterManager{
		peers: make(map[string]*Peer),
		eng:   engine,
	}
}

// AddPeer adds a peer to the cluster
func (cm *ClusterManager) AddPeer(id, addr string) error {
	if !cm.eng.IsLeader() {
		return fmt.Errorf("only leader can add peers")
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.peers[id]; exists {
		return fmt.Errorf("peer %s already exists", id)
	}

	// Add peer to Raft cluster
	if err := cm.eng.AddPeer(id, addr); err != nil {
		return err
	}

	cm.peers[id] = &Peer{
		ID:       id,
		Addr:     addr,
		Status:   "alive",
		LastSeen: time.Now(),
	}

	return nil
}

// RemovePeer removes a peer from the cluster
func (cm *ClusterManager) RemovePeer(id string) error {
	if !cm.eng.IsLeader() {
		return fmt.Errorf("only leader can remove peers")
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.peers[id]; !exists {
		return fmt.Errorf("peer %s not found", id)
	}

	// Remove peer from Raft cluster
	if err := cm.eng.RemovePeer(id); err != nil {
		return err
	}

	delete(cm.peers, id)
	return nil
}

// GetPeers returns all peers
func (cm *ClusterManager) GetPeers() map[string]*Peer {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make(map[string]*Peer)
	for k, v := range cm.peers {
		result[k] = v
	}
	return result
}

// UpdatePeerStatus updates the status of a peer
func (cm *ClusterManager) UpdatePeerStatus(id, status string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	peer, exists := cm.peers[id]
	if !exists {
		return fmt.Errorf("peer %s not found", id)
	}

	peer.Status = status
	peer.LastSeen = time.Now()
	return nil
}
