package commit

import (
	"fmt"
	"sort"

	"github.com/erigontech/erigon/common"
	"github.com/erigontech/erigon/execution/core/parallel/mvcc"
	"github.com/erigontech/erigon/execution/types"
)

// CommitManager handles committing parallel execution results to state
type CommitManager interface {
	// Commit applies all transaction results to the state in order
	// state should be parallel.StateDB, result should be *parallel.ExecutionResult
	Commit(results interface{}, state interface{}) (interface{}, error)

	// Reset clears commit manager state for a new block
	Reset()
}

// StateDB interface (duplicated to avoid import cycle)
type StateDB interface {
	SetAccount(addr common.Address, account interface{}) error
	SetState(addr common.Address, key, value common.Hash) error
	SetCode(addr common.Address, code []byte) error
	SetBalance(addr common.Address, balance interface{}) error
	SetNonce(addr common.Address, nonce uint64) error
}

// manager implements CommitManager
type manager struct {
	committedTxs int
}

// NewCommitManager creates a new commit manager
func NewCommitManager() CommitManager {
	return &manager{
		committedTxs: 0,
	}
}

// Commit applies all transaction results to the state in order
func (m *manager) Commit(results interface{}, state interface{}) (interface{}, error) {
	// Cast state to StateDB
	stateDB, ok := state.(StateDB)
	if !ok {
		return nil, fmt.Errorf("invalid state type")
	}
	
	// Cast results to []*TxResult
	txResults, ok := results.([]*TxResult)
	if !ok {
		return nil, fmt.Errorf("invalid results type")
	}
	
	if len(txResults) == 0 {
		return &ExecutionResult{
			Receipts:  make([]*types.Receipt, 0),
			Logs:      make([]*types.Log, 0),
			GasUsed:   0,
			StateRoot: common.Hash{},
		}, nil
	}

	// Sort results by transaction index to ensure correct order
	sortedResults := make([]*TxResult, len(txResults))
	copy(sortedResults, txResults)
	sort.Slice(sortedResults, func(i, j int) bool {
		return sortedResults[i].TxIndex < sortedResults[j].TxIndex
	})

	// Aggregate results
	var receipts []*types.Receipt
	var logs []*types.Log
	var totalGasUsed uint64
	var totalBlobGasUsed uint64

	for _, result := range sortedResults {
		if result.Error != nil {
			return nil, fmt.Errorf("cannot commit transaction %d with error: %w", result.TxIndex, result.Error)
		}

		// Add receipt
		if result.Receipt != nil {
			receipts = append(receipts, result.Receipt)
		}

		// Add logs
		if result.Logs != nil {
			logs = append(logs, result.Logs...)
		}

		// Accumulate gas
		totalGasUsed += result.GasUsed
		totalBlobGasUsed += result.BlobGasUsed

		// Apply write set to state
		if err := m.applyWriteSet(result, stateDB); err != nil {
			return nil, fmt.Errorf("failed to apply write set for tx %d: %w", result.TxIndex, err)
		}

		m.committedTxs++
	}

	// Calculate state root (placeholder for now)
	stateRoot := common.Hash{}

	return &ExecutionResult{
		Receipts:    receipts,
		Logs:        logs,
		GasUsed:     totalGasUsed,
		BlobGasUsed: totalBlobGasUsed,
		StateRoot:   stateRoot,
	}, nil
}

// ExecutionResult contains the results of block execution
type ExecutionResult struct {
	Receipts    []*types.Receipt
	Logs        []*types.Log
	GasUsed     uint64
	BlobGasUsed uint64
	StateRoot   common.Hash
}

// applyWriteSet applies a transaction's write set to the state
func (m *manager) applyWriteSet(result *TxResult, state StateDB) error {
	if result.WriteSet == nil {
		return nil
	}

	// Cast WriteSet to mvcc.WriteSet
	writeSet, ok := result.WriteSet.(*mvcc.WriteSet)
	if !ok {
		return fmt.Errorf("invalid write set type")
	}

	// Apply account writes
	for addr, account := range writeSet.Accounts {
		if err := state.SetAccount(addr, account); err != nil {
			return fmt.Errorf("failed to set account %s: %w", addr.Hex(), err)
		}
	}

	// Apply storage writes
	for storageKey, value := range writeSet.Storage {
		valueHash := common.BytesToHash(value.Bytes())
		if err := state.SetState(storageKey.Address, storageKey.Slot, valueHash); err != nil {
			return fmt.Errorf("failed to set storage: %w", err)
		}
	}

	// Apply code writes
	for addr, code := range writeSet.Code {
		if err := state.SetCode(addr, code); err != nil {
			return fmt.Errorf("failed to set code for %s: %w", addr.Hex(), err)
		}
	}

	// Apply balance writes
	for addr, balance := range writeSet.Balances {
		if err := state.SetBalance(addr, balance); err != nil {
			return fmt.Errorf("failed to set balance for %s: %w", addr.Hex(), err)
		}
	}

	// Apply nonce writes
	for addr, nonce := range writeSet.Nonces {
		if err := state.SetNonce(addr, nonce); err != nil {
			return fmt.Errorf("failed to set nonce for %s: %w", addr.Hex(), err)
		}
	}

	return nil
}

// Reset clears commit manager state for a new block
func (m *manager) Reset() {
	m.committedTxs = 0
}

// GetCommittedCount returns the number of committed transactions
func (m *manager) GetCommittedCount() int {
	return m.committedTxs
}

// TxResult represents the result of executing a transaction
// Defined here to avoid import cycles
type TxResult struct{
	TxIndex     int
	Receipt     *types.Receipt
	Logs        []*types.Log
	GasUsed     uint64
	BlobGasUsed uint64
	WriteSet    interface{} // *mvcc.WriteSet
	Error       error
}


