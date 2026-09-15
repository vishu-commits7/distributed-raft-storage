package tests

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/raft"
	"github.com/stretchr/testify/assert"
	"github.com/vishu-commits7/distributed-raft-storage/pkg"
)

func TestStateMachineSet(t *testing.T) {
	sm := pkg.NewStateMachine()

	// Create a set command
	entry := pkg.LogEntry{
		Op:    "set",
		Key:   "test",
		Value: "value",
	}
	b, _ := json.Marshal(entry)

	// Apply the command
	log := &raft.Log{
		Index: 1,
		Term:  1,
		Type:  raft.LogCommand,
		Data:  b,
	}

	result := sm.Apply(log)
	assert.Nil(t, result)

	// Verify the value was set
	val, ok := sm.Get("test")
	assert.True(t, ok)
	assert.Equal(t, "value", val)
}

func TestStateMachineDelete(t *testing.T) {
	sm := pkg.NewStateMachine()

	// Set a value first
	entry := pkg.LogEntry{
		Op:    "set",
		Key:   "test",
		Value: "value",
	}
	b, _ := json.Marshal(entry)
	log := &raft.Log{Index: 1, Term: 1, Type: raft.LogCommand, Data: b}
	sm.Apply(log)

	// Delete the value
	deleteEntry := pkg.LogEntry{
		Op:  "delete",
		Key: "test",
	}
	b, _ = json.Marshal(deleteEntry)
	log = &raft.Log{Index: 2, Term: 1, Type: raft.LogCommand, Data: b}
	result := sm.Apply(log)
	assert.Nil(t, result)

	// Verify the value was deleted
	_, ok := sm.Get("test")
	assert.False(t, ok)
}

func TestStateMachineScan(t *testing.T) {
	sm := pkg.NewStateMachine()

	// Set multiple values
	for i := 0; i < 5; i++ {
		entry := pkg.LogEntry{
			Op:    "set",
			Key:   "key" + string(rune(48+i)),
			Value: "value" + string(rune(48+i)),
		}
		b, _ := json.Marshal(entry)
		log := &raft.Log{Index: int64(i + 1), Term: 1, Type: raft.LogCommand, Data: b}
		sm.Apply(log)
	}

	// Scan with prefix
	results := sm.Scan("key")
	assert.Equal(t, 5, len(results))
}
