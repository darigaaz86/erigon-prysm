package parallel

import (
	"github.com/erigontech/erigon/common"
	"github.com/erigontech/erigon/execution/types"
	"github.com/erigontech/erigon/execution/types/accounts"
	"github.com/holiman/uint256"
)

// ParallelExecutor orchestrates parallel execution of transactions in a block.
type ParallelExecutor interface {
	// ExecuteBlock executes all transactions in a block in parallel
	// Takes header, transactions, and state instead of a Block object
	ExecuteBlock(header *types.Header, txs []types.Transaction, state StateDB) (*ExecutionResult, error)

	// SetWorkerCount configures the number of worker threads
	SetWorkerCount(count int)

	// GetMetrics returns execution metrics
	GetMetrics() *ExecutionMetrics
}

// StateDB provides access to blockchain state using Erigon's production types.
// This interface is designed to work with SharedDomains from db/state.
type StateDB interface {
	// Account operations
	GetAccount(addr common.Address) (*accounts.Account, error)
	SetAccount(addr common.Address, account *accounts.Account) error
	
	// Storage operations
	GetState(addr common.Address, key common.Hash) (common.Hash, error)
	SetState(addr common.Address, key, value common.Hash) error
	
	// Code operations
	GetCode(addr common.Address) ([]byte, error)
	SetCode(addr common.Address, code []byte) error
	
	// Balance operations (convenience methods)
	GetBalance(addr common.Address) (*uint256.Int, error)
	SetBalance(addr common.Address, balance *uint256.Int) error
	
	// Nonce operations (convenience methods)
	GetNonce(addr common.Address) (uint64, error)
	SetNonce(addr common.Address, nonce uint64) error
}

// Transaction wraps execution/types.Transaction with additional methods needed for parallel execution
type Transaction interface {
	// Core transaction data
	Hash() common.Hash
	Type() byte
	
	// Sender and recipient
	GetSender() (common.Address, error)
	GetTo() *common.Address
	
	// Value and data
	GetValue() *uint256.Int
	GetData() []byte
	
	// Gas parameters
	GetGas() uint64
	GetPrice() *uint256.Int
	GetNonce() uint64
	
	// Chain parameters
	GetChainID() *uint256.Int
	
	// Transaction properties
	IsContractCreation() bool
	Protected() bool
	
	// Access list (EIP-2930)
	GetAccessList() types.AccessList
	
	// Blob transactions (EIP-4844)
	GetBlobHashes() []common.Hash
	GetBlobGas() uint64
}

// TransactionWrapper wraps execution/types.Transaction to implement the Transaction interface
type TransactionWrapper struct {
	tx     types.Transaction
	sender common.Address
	index  int
}

// NewTransactionWrapper creates a new transaction wrapper
func NewTransactionWrapper(tx types.Transaction, sender common.Address, index int) Transaction {
	return &TransactionWrapper{
		tx:     tx,
		sender: sender,
		index:  index,
	}
}

// Implement Transaction interface
func (tw *TransactionWrapper) Hash() common.Hash {
	return tw.tx.Hash()
}

func (tw *TransactionWrapper) Type() byte {
	return tw.tx.Type()
}

func (tw *TransactionWrapper) GetSender() (common.Address, error) {
	return tw.sender, nil
}

func (tw *TransactionWrapper) GetTo() *common.Address {
	return tw.tx.GetTo()
}

func (tw *TransactionWrapper) GetValue() *uint256.Int {
	return tw.tx.GetValue()
}

func (tw *TransactionWrapper) GetData() []byte {
	return tw.tx.GetData()
}

func (tw *TransactionWrapper) GetGas() uint64 {
	return tw.tx.GetGasLimit()
}

func (tw *TransactionWrapper) GetPrice() *uint256.Int {
	// For execution/types.Transaction, use GetFeeCap or GetTipCap
	// This is a simplified implementation - use fee cap
	return tw.tx.GetFeeCap()
}

func (tw *TransactionWrapper) GetNonce() uint64 {
	return tw.tx.GetNonce()
}

func (tw *TransactionWrapper) GetChainID() *uint256.Int {
	return tw.tx.GetChainID()
}

func (tw *TransactionWrapper) IsContractCreation() bool {
	return tw.tx.GetTo() == nil
}

func (tw *TransactionWrapper) Protected() bool {
	return tw.tx.Protected()
}

func (tw *TransactionWrapper) GetAccessList() types.AccessList {
	return tw.tx.GetAccessList()
}

func (tw *TransactionWrapper) GetBlobHashes() []common.Hash {
	return tw.tx.GetBlobHashes()
}

func (tw *TransactionWrapper) GetBlobGas() uint64 {
	return tw.tx.GetBlobGas()
}
