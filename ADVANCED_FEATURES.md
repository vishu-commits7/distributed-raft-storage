# Advanced Features

The Distributed Raft Storage Engine includes several advanced features for production deployments:

## 1. Audit Logging

Track all operations for compliance and debugging:

```go
auditLog := pkg.NewAuditLog(10000)

auditLog.Log(pkg.AuditEntry{
    NodeID:    "node1",
    Operation: "set",
    Key:       "key1",
    NewValue:  "value1",
    Success:   true,
})

// Query audit log
entries := auditLog.GetEntriesByKey("key1")
entries := auditLog.GetEntriesByOperation("set")
entries := auditLog.GetEntriesSince(time.Now().Add(-1*time.Hour))

// Export to JSON
json, _ := auditLog.Export()
```

## 2. Snapshot Management

Create and verify snapshots:

```go
snapMgr := pkg.NewSnapshotManager(engine)

// Create snapshot
snapshot, _ := snapMgr.Create("snapshot-2024-01-01")

// Verify integrity
isValid, _ := snapMgr.Verify("snapshot-2024-01-01")

// List and delete
ids := snapMgr.List()
snapMgr.Delete("snapshot-2024-01-01")
```

## 3. Backup Management

Handle backup and recovery:

```go
backupMgr := pkg.NewBackupManager(engine, auditLog)

// Create backup
backup, _ := backupMgr.CreateBackup("backup-2024-01-01")

// Monitor progress
status := backupMgr.GetBackup("backup-2024-01-01")
fmt.Printf("Status: %s, Count: %d\n", status.Status, status.DataCount)

// List and delete
backups := backupMgr.ListBackups()
backupMgr.DeleteBackup("backup-2024-01-01")
```

## 4. Caching Layer

Improve performance with TTL-based caching:

```go
cache := pkg.NewCache(1000, 5*time.Minute)

// Set with default TTL
cache.Set("key", "value")

// Set with custom TTL
cache.SetWithTTL("key", "value", 10*time.Minute)

// Get
value, found := cache.Get("key")

// Cleanup expired entries
expiredCount := cache.CleanupExpired()
```

## 5. Transaction Support

Support ACID transactions:

```go
txlog := pkg.NewTransactionLog()

// Begin transaction
tx := txlog.Begin("tx-123")

// Add operations
txlog.AddOp("tx-123", pkg.TransactionOp{
    Op:       "set",
    Key:      "key1",
    NewValue: "value1",
})

// Commit or rollback
txlog.Commit("tx-123")
// txlog.Rollback("tx-123")

// Query
pendingTxs := txlog.GetPending()
```

## 6. Rate Limiting

Protect against abuse:

```go
rateLimiter := pkg.NewRateLimiter()

// Set limits per client
rateLimiter.SetLimit("api-client-1", 100) // 100 req/sec
rateLimiter.SetLimit("api-client-2", 50)  // 50 req/sec

// Check if request allowed
if rateLimiter.Allow("api-client-1") {
    // Process request
} else {
    // Return 429 Too Many Requests
}

// Monitor status
status := rateLimiter.GetStatus("api-client-1")
fmt.Printf("Throttled: %v\n", status.IsThrottled)
```

## 7. Health Checking

Monitor node health:

```go
healthChecker := pkg.NewHealthChecker(engine)
healthChecker.Start(5 * time.Second) // Check every 5 seconds

// Get status
status := healthChecker.GetStatus()
fmt.Printf("Healthy: %v, State: %s\n", status.IsHealthy, status.RaftState)

healthChecker.Stop()
```

## 8. Replication Management

Manage cluster replication:

```go
replMgr := pkg.NewReplicationManager(engine)

// Sync a peer
replMgr.SyncPeer("node2")

// Get peer status
status := replMgr.GetPeerStatus("node2")
fmt.Printf("In sync: %v, Lag: %d ms\n", status.IsInSync, status.SyncLag)

// Get all peer status
allStatus := replMgr.GetAllPeerStatus()

// Get replication events
events := replMgr.GetReplicationEvents(10)
```

## Integration Example

```go
engine, _ := pkg.NewRaftEngine(config)

// Setup components
auditLog := pkg.NewAuditLog(10000)
cache := pkg.NewCache(1000, 5*time.Minute)
snapMgr := pkg.NewSnapshotManager(engine)
backupMgr := pkg.NewBackupManager(engine, auditLog)
healthChecker := pkg.NewHealthChecker(engine)
rateLimiter := pkg.NewRateLimiter()

// Start health checks
healthChecker.Start(10 * time.Second)

// Process request with all features
clientID := "client-123"
if !rateLimiter.Allow(clientID) {
    return fmt.Errorf("rate limited")
}

value, found := cache.Get("key")
if !found {
    value, found = engine.Get("key")
    if found {
        cache.Set("key", value)
    }
}

auditLog.Log(pkg.AuditEntry{
    NodeID:    engine.config.NodeID,
    Operation: "get",
    Key:       "key",
    Success:   found,
    ActorIP:   clientID,
})

return value, found
```
