package validator

import (
	"github.com/erigontech/erigon/common"
	"github.com/erigontech/erigon/execution/core/parallel/mvcc"
)

// ConflictValidator detects conflicts between transactions
type ConflictValidator interface {
	// Validate checks all transactions for conflicts
	// Returns indices of transactions that need re-execution
	Validate(results []interface{}) []int

	// HasConflict checks if two transactions conflict
	HasConflict(tx1, tx2 interface{}) bool

	// Reset clears validator state for a new block
	Reset()
}

// validator implements ConflictValidator
type validator struct {
	conflictCount int
}

// NewConflictValidator creates a new conflict validator
func NewConflictValidator() ConflictValidator {
	return &validator{
		conflictCount: 0,
	}
}

// Validate checks all transactions for conflicts
func (v *validator) Validate(results []interface{}) []int {
	var invalidTxs []int

	// Check each transaction against all previous transactions
	for i := 1; i < len(results); i++ {
		currentTx, ok := results[i].(*TxResult)
		if !ok {
			continue
		}

		// Check against all previous transactions
		for j := 0; j < i; j++ {
			prevTx, ok := results[j].(*TxResult)
			if !ok {
				continue
			}

			// Check if current transaction's read set conflicts with previous transaction's write set
			if currentTx.ReadSet != nil && prevTx.WriteSet != nil {
				if v.hasReadWriteConflict(currentTx.ReadSet, prevTx.WriteSet) {
					invalidTxs = append(invalidTxs, i)
					v.conflictCount++
					break // No need to check further
				}
			}
		}
	}

	return invalidTxs
}

// HasConflict checks if two transactions conflict
func (v *validator) HasConflict(tx1, tx2 interface{}) bool {
	result1, ok1 := tx1.(*TxResult)
	result2, ok2 := tx2.(*TxResult)
	if !ok1 || !ok2 {
		return false
	}
	
	if result2.ReadSet == nil || result1.WriteSet == nil {
		return false
	}
	
	return v.hasReadWriteConflict(result2.ReadSet, result1.WriteSet)
}

// hasReadWriteConflict checks if a read set conflicts with a write set
func (v *validator) hasReadWriteConflict(readSet *mvcc.ReadSet, writeSet *mvcc.WriteSet) bool {
	if readSet == nil || writeSet == nil {
		return false
	}

	// Check account conflicts
	for addr := range readSet.Accounts {
		if _, exists := writeSet.Accounts[addr]; exists {
			return true
		}
	}

	// Check storage conflicts
	for storageKey := range readSet.Storage {
		if _, exists := writeSet.Storage[storageKey]; exists {
			return true
		}
	}

	// Check code conflicts
	for addr := range readSet.Code {
		if _, exists := writeSet.Code[addr]; exists {
			return true
		}
	}

	// Check balance conflicts
	for addr := range readSet.Balances {
		if _, exists := writeSet.Balances[addr]; exists {
			return true
		}
	}

	// Check nonce conflicts
	for addr := range readSet.Nonces {
		if _, exists := writeSet.Nonces[addr]; exists {
			return true
		}
	}

	return false
}

// Reset clears validator state for a new block
func (v *validator) Reset() {
	v.conflictCount = 0
}

// GetConflictCount returns the total number of conflicts detected
func (v *validator) GetConflictCount() int {
	return v.conflictCount
}

// TxResult represents the result of executing a transaction
type TxResult struct {
	TxIndex     int
	TxHash      common.Hash
	Receipt     interface{} // Placeholder
	Logs        interface{} // Placeholder
	GasUsed     uint64
	BlobGasUsed uint64
	ReadSet     *mvcc.ReadSet
	WriteSet    *mvcc.WriteSet
	Error       error
}

// BuildDependencyGraph builds a dependency graph from transaction results
func BuildDependencyGraph(results []*TxResult) map[int][]int {
	dependencies := make(map[int][]int)

	for i := 1; i < len(results); i++ {
		currentTx := results[i]

		// Find all transactions that this transaction depends on
		for j := 0; j < i; j++ {
			prevTx := results[j]

			// If current tx reads what previous tx wrote, there's a dependency
			if hasReadWriteConflict(currentTx.ReadSet, prevTx.WriteSet) {
				dependencies[i] = append(dependencies[i], j)
			}
		}
	}

	return dependencies
}

// hasReadWriteConflict is a helper function for dependency graph building
func hasReadWriteConflict(readSet *mvcc.ReadSet, writeSet *mvcc.WriteSet) bool {
	if readSet == nil || writeSet == nil {
		return false
	}

	// Check account conflicts
	for addr := range readSet.Accounts {
		if _, exists := writeSet.Accounts[addr]; exists {
			return true
		}
	}

	// Check storage conflicts
	for storageKey := range readSet.Storage {
		if _, exists := writeSet.Storage[storageKey]; exists {
			return true
		}
	}

	return false
}
