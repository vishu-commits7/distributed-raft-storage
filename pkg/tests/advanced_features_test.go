package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vishu-commits7/distributed-raft-storage/pkg"
)

func TestAuditLog(t *testing.T) {
	al := pkg.NewAuditLog(1000)

	// Log an entry
	al.Log(pkg.AuditEntry{
		NodeID:    "node1",
		Operation: "set",
		Key:       "key1",
		NewValue:  "value1",
		Success:   true,
	})

	// Verify entry was logged
	entries := al.GetEntries()
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "set", entries[0].Operation)
}

func TestCache(t *testing.T) {
	cache := pkg.NewCache(10, 0) // unlimited

	// Set and get
	cache.Set("key", "value")
	val, ok := cache.Get("key")

	assert.True(t, ok)
	assert.Equal(t, "value", val)
}

func TestTransactionLog(t *testing.T) {
	txlog := pkg.NewTransactionLog()

	// Begin transaction
	tx := txlog.Begin("tx1")
	assert.Equal(t, "pending", tx.Status)

	// Add operation
	opErr := txlog.AddOp("tx1", pkg.TransactionOp{
		Op:       "set",
		Key:      "key1",
		NewValue: "value1",
	})
	assert.NoError(t, opErr)

	// Commit
	commitErr := txlog.Commit("tx1")
	assert.NoError(t, commitErr)

	// Verify
	tx = txlog.Get("tx1")
	assert.Equal(t, "committed", tx.Status)
}

func TestRateLimiter(t *testing.T) {
	rl := pkg.NewRateLimiter()

	// Set limit for client
	rl.SetLimit("client1", 5)

	// Check validation
	isValid := rl.ValidateClientID("client1")
	assert.True(t, isValid)

	isValid = rl.ValidateClientID("invalid@client")
	assert.False(t, isValid)
}
