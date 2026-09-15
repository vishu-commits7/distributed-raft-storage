package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vishu-commits7/distributed-raft-storage/pkg"
)

func TestRaftEngineCreation(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &pkg.RaftConfig{
		NodeID:    "test-node",
		DataDir:   tempDir,
		BindAddr:  "127.0.0.1",
		BindPort:  6379,
		Bootstrap: true,
	}

	eng, err := pkg.NewRaftEngine(cfg)
	assert.NoError(t, err)
	defer eng.Close()

	// Give it a moment to become leader
	time.Sleep(100 * time.Millisecond)

	assert.True(t, eng.IsLeader())
}

func TestRaftEngineSetAndGet(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &pkg.RaftConfig{
		NodeID:    "test-node",
		DataDir:   tempDir,
		BindAddr:  "127.0.0.1",
		BindPort:  6380,
		Bootstrap: true,
	}

	eng, err := pkg.NewRaftEngine(cfg)
	assert.NoError(t, err)
	defer eng.Close()

	// Wait for leadership
	time.Sleep(100 * time.Millisecond)

	// Set a value
	err = eng.Set("key", "value")
	assert.NoError(t, err)

	// Get the value
	val, ok := eng.Get("key")
	assert.True(t, ok)
	assert.Equal(t, "value", val)
}

func TestRaftEngineDelete(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &pkg.RaftConfig{
		NodeID:    "test-node",
		DataDir:   tempDir,
		BindAddr:  "127.0.0.1",
		BindPort:  6381,
		Bootstrap: true,
	}

	eng, err := pkg.NewRaftEngine(cfg)
	assert.NoError(t, err)
	defer eng.Close()

	// Wait for leadership
	time.Sleep(100 * time.Millisecond)

	// Set a value
	eng.Set("key", "value")

	// Delete it
	err = eng.Delete("key")
	assert.NoError(t, err)

	// Verify it's gone
	_, ok := eng.Get("key")
	assert.False(t, ok)
}

func TestRaftEngineScan(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &pkg.RaftConfig{
		NodeID:    "test-node",
		DataDir:   tempDir,
		BindAddr:  "127.0.0.1",
		BindPort:  6382,
		Bootstrap: true,
	}

	eng, err := pkg.NewRaftEngine(cfg)
	assert.NoError(t, err)
	defer eng.Close()

	// Wait for leadership
	time.Sleep(100 * time.Millisecond)

	// Set multiple values
	for i := 0; i < 5; i++ {
		key := "key:" + string(rune(48+i))
		val := "value:" + string(rune(48+i))
		eng.Set(key, val)
	}

	// Scan
	results := eng.Scan("key:")
	assert.Equal(t, 5, len(results))
}
