package parallel

import (
	"time"

	"github.com/erigontech/erigon/common"
	"github.com/erigontech/erigon/execution/types"
	"github.com/erigontech/erigon/execution/types/accounts"
	"github.com/holiman/uint256"
)

// ExecutionResult contains the results of block execution using production types.
type ExecutionResult struct {
	// Transaction receipts
	Receipts []*types.Receipt
	
	// Gas accounting
	GasUsed     uint64
	BlobGasUsed uint64
	
	// State root after execution
	StateRoot common.Hash
	
	// Event logs from all transactions
	Logs []*types.Log
	
	// Parallel execution metrics
	Metrics *ExecutionMetrics
}

// ExecutionMetrics tracks performance metrics for parallel execution.
type ExecutionMetrics struct {
	// Transaction counts
	TotalTxs      int
	ParallelTxs   int
	ReexecutedTxs int
	
	// Conflict statistics
	ConflictCount int
	ConflictRate  float64
	
	// Timing information
	ExecutionTime   time.Duration
	ValidationTime  time.Duration
	ReexecutionTime time.Duration
	CommitTime      time.Duration
	TotalDuration   time.Duration
	
	// Performance metrics
	SpeedupFactor      float64
	ParallelEfficiency float64
	
	// Re-execution statistics
	ReexecutionCount int
	MaxReexecutions  int
	AvgReexecutions  float64
}

// NewExecutionMetrics creates a new ExecutionMetrics instance
func NewExecutionMetrics() *ExecutionMetrics {
	return &ExecutionMetrics{
		TotalTxs:      0,
		ParallelTxs:   0,
		ReexecutedTxs: 0,
		ConflictCount: 0,
		ConflictRate:  0.0,
		SpeedupFactor: 1.0,
	}
}



// ReadSet tracks all state reads performed by a transaction
type ReadSet struct {
	// Account reads: address -> account data
	Accounts map[common.Address]*accounts.Account
	
	// Storage reads: address -> (key -> value)
	Storage map[common.Address]map[common.Hash]common.Hash
	
	// Code reads: address -> code hash
	Code map[common.Address]common.Hash
	
	// Balance reads: address -> balance
	Balances map[common.Address]*uint256.Int
	
	// Nonce reads: address -> nonce
	Nonces map[common.Address]uint64
}

// NewReadSet creates a new ReadSet
func NewReadSet() *ReadSet {
	return &ReadSet{
		Accounts: make(map[common.Address]*accounts.Account),
		Storage:  make(map[common.Address]map[common.Hash]common.Hash),
		Code:     make(map[common.Address]common.Hash),
		Balances: make(map[common.Address]*uint256.Int),
		Nonces:   make(map[common.Address]uint64),
	}
}

// WriteSet tracks all state writes performed by a transaction
type WriteSet struct {
	// Account writes: address -> account data
	Accounts map[common.Address]*accounts.Account
	
	// Storage writes: address -> (key -> value)
	Storage map[common.Address]map[common.Hash]common.Hash
	
	// Code writes: address -> code
	Code map[common.Address][]byte
	
	// Balance writes: address -> balance
	Balances map[common.Address]*uint256.Int
	
	// Nonce writes: address -> nonce
	Nonces map[common.Address]uint64
	
	// Created contracts
	CreatedContracts []common.Address
	
	// Deleted accounts (SELFDESTRUCT)
	DeletedAccounts []common.Address
}

// NewWriteSet creates a new WriteSet
func NewWriteSet() *WriteSet {
	return &WriteSet{
		Accounts:         make(map[common.Address]*accounts.Account),
		Storage:          make(map[common.Address]map[common.Hash]common.Hash),
		Code:             make(map[common.Address][]byte),
		Balances:         make(map[common.Address]*uint256.Int),
		Nonces:           make(map[common.Address]uint64),
		CreatedContracts: make([]common.Address, 0),
		DeletedAccounts:  make([]common.Address, 0),
	}
}

// HasConflict checks if this write set conflicts with a read set
func (ws *WriteSet) HasConflict(rs *ReadSet) bool {
	// Check account conflicts
	for addr := range ws.Accounts {
		if _, exists := rs.Accounts[addr]; exists {
			return true
		}
	}
	
	// Check storage conflicts
	for addr, writes := range ws.Storage {
		if reads, exists := rs.Storage[addr]; exists {
			for key := range writes {
				if _, readExists := reads[key]; readExists {
					return true
				}
			}
		}
	}
	
	// Check code conflicts
	for addr := range ws.Code {
		if _, exists := rs.Code[addr]; exists {
			return true
		}
	}
	
	// Check balance conflicts
	for addr := range ws.Balances {
		if _, exists := rs.Balances[addr]; exists {
			return true
		}
	}
	
	// Check nonce conflicts
	for addr := range ws.Nonces {
		if _, exists := rs.Nonces[addr]; exists {
			return true
		}
	}
	
	return false
}

// TxResult contains the result of executing a single transaction (shared type)
type TxResult struct {
	// Transaction identification
	TxIndex int
	TxHash  common.Hash

	// Execution results
	Receipt *types.Receipt
	Logs    []*types.Log

	// Gas accounting
	GasUsed     uint64
	BlobGasUsed uint64

	// State tracking (forward declarations to avoid import cycles)
	ReadSet  interface{} // Will be *mvcc.ReadSet
	WriteSet interface{} // Will be *mvcc.WriteSet

	// Re-execution tracking
	ReexecCount int

	// Error if execution failed
	Error error
}
