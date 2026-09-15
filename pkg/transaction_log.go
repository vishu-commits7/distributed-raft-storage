package pkg

import (
	"fmt"
	"sync"
)

// TransactionLog tracks transactions for ACID properties
type TransactionLog struct {
	mu           sync.RWMutex
	transactions map[string]*Transaction
	logStack     []string // Stack of transaction IDs
}

// Transaction represents a single transaction
type Transaction struct {
	ID      string
	Status  string // "pending", "committed", "rolled_back"
	Ops     []TransactionOp
	Started int64
}

// TransactionOp represents an operation within a transaction
type TransactionOp struct {
	Op       string // "set", "delete"
	Key      string
	OldValue string
	NewValue string
}

// NewTransactionLog creates a new transaction log
func NewTransactionLog() *TransactionLog {
	return &TransactionLog{
		transactions: make(map[string]*Transaction),
		logStack:     make([]string, 0),
	}
}

// Begin starts a new transaction
func (tl *TransactionLog) Begin(id string) *Transaction {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	tx := &Transaction{
		ID:      id,
		Status:  "pending",
		Ops:     make([]TransactionOp, 0),
		Started: int64(0),
	}

	tl.transactions[id] = tx
	tl.logStack = append(tl.logStack, id)

	return tx
}

// AddOp adds an operation to a transaction
func (tl *TransactionLog) AddOp(txID string, op TransactionOp) error {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	tx, ok := tl.transactions[txID]
	if !ok {
		return fmt.Errorf("transaction not found: %s", txID)
	}

	if tx.Status != "pending" {
		return fmt.Errorf("transaction is not pending: %s", txID)
	}

	tx.Ops = append(tx.Ops, op)
	return nil
}

// Commit marks a transaction as committed
func (tl *TransactionLog) Commit(txID string) error {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	tx, ok := tl.transactions[txID]
	if !ok {
		return fmt.Errorf("transaction not found: %s", txID)
	}

	tx.Status = "committed"
	return nil
}

// Rollback marks a transaction as rolled back
func (tl *TransactionLog) Rollback(txID string) error {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	tx, ok := tl.transactions[txID]
	if !ok {
		return fmt.Errorf("transaction not found: %s", txID)
	}

	tx.Status = "rolled_back"
	return nil
}

// Get retrieves a transaction
func (tl *TransactionLog) Get(txID string) *Transaction {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	if tx, ok := tl.transactions[txID]; ok {
		copy := *tx
		copy.Ops = make([]TransactionOp, len(tx.Ops))
		copy(copy.Ops, tx.Ops)
		return &copy
	}
	return nil
}

// GetPending returns all pending transactions
func (tl *TransactionLog) GetPending() []*Transaction {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	var pending []*Transaction
	for _, tx := range tl.transactions {
		if tx.Status == "pending" {
			copy := *tx
			copy.Ops = make([]TransactionOp, len(tx.Ops))
			copy(copy.Ops, tx.Ops)
			pending = append(pending, &copy)
		}
	}
	return pending
}

// Size returns the number of transactions
func (tl *TransactionLog) Size() int {
	tl.mu.RLock()
	defer tl.mu.RUnlock()
	return len(tl.transactions)
}
