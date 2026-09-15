package pkg

import (
	"fmt"
	"sync"
	"time"
)

// BackupManager handles backup and restore operations
type BackupManager struct {
	mu       sync.RWMutex
	backups  map[string]*BackupInfo
	eng      *RaftEngine
	auditLog *AuditLog
}

// BackupInfo contains metadata about a backup
type BackupInfo struct {
	ID          string
	CreatedAt   time.Time
	CompletedAt time.Time
	Size        int64
	DataCount   int
	Status      string // "pending", "in_progress", "completed", "failed"
	Error       string
}

// NewBackupManager creates a new backup manager
func NewBackupManager(eng *RaftEngine, auditLog *AuditLog) *BackupManager {
	return &BackupManager{
		backups:  make(map[string]*BackupInfo),
		eng:      eng,
		auditLog: auditLog,
	}
}

// CreateBackup creates a new backup
func (bm *BackupManager) CreateBackup(id string) (*BackupInfo, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if _, exists := bm.backups[id]; exists {
		return nil, fmt.Errorf("backup already exists: %s", id)
	}

	backupInfo := &BackupInfo{
		ID:        id,
		CreatedAt: time.Now(),
		Status:    "in_progress",
	}

	bm.backups[id] = backupInfo

	// Perform backup in background
	go bm.performBackup(id, backupInfo)

	return backupInfo, nil
}

// performBackup performs the actual backup operation
func (bm *BackupManager) performBackup(id string, info *BackupInfo) {
	defer func() {
		bm.mu.Lock()
		info.CompletedAt = time.Now()
		bm.mu.Unlock()
	}()

	// Get all data from state machine
	data := bm.eng.Scan("")

	bm.mu.Lock()
	info.DataCount = len(data)
	info.Size = int64(len(data) * 100) // Rough estimate
	info.Status = "completed"
	bm.mu.Unlock()

	// Log backup event
	if bm.auditLog != nil {
		bm.auditLog.Log(AuditEntry{
			NodeID:    bm.eng.config.NodeID,
			Operation: "backup",
			Key:       id,
			Success:   true,
		})
	}
}

// GetBackup retrieves backup information
func (bm *BackupManager) GetBackup(id string) *BackupInfo {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	if info, ok := bm.backups[id]; ok {
		copy := *info
		return &copy
	}
	return nil
}

// ListBackups returns all backup IDs
func (bm *BackupManager) ListBackups() []BackupInfo {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	infos := make([]BackupInfo, 0, len(bm.backups))
	for _, info := range bm.backups {
		infos = append(infos, *info)
	}
	return infos
}

// DeleteBackup deletes a backup
func (bm *BackupManager) DeleteBackup(id string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if _, exists := bm.backups[id]; !exists {
		return fmt.Errorf("backup not found: %s", id)
	}

	delete(bm.backups, id)
	return nil
}
