package mvcc

import (
	"fmt"

	"github.com/erigontech/erigon/common"
	"github.com/erigontech/erigon/execution/types/accounts"
	"github.com/holiman/uint256"
)

// StateDB interface (duplicated to avoid import cycle)
type StateDB interface {
	GetAccount(addr common.Address) (*accounts.Account, error)
	SetAccount(addr common.Address, account *accounts.Account) error
	GetState(addr common.Address, key common.Hash) (common.Hash, error)
	SetState(addr common.Address, key, value common.Hash) error
	GetCode(addr common.Address) ([]byte, error)
	SetCode(addr common.Address, code []byte) error
	GetBalance(addr common.Address) (*uint256.Int, error)
	SetBalance(addr common.Address, balance *uint256.Int) error
	GetNonce(addr common.Address) (uint64, error)
	SetNonce(addr common.Address, nonce uint64) error
}

// StateWrapper wraps an MVCC state manager to implement the StateDB interface
type StateWrapper struct {
	mvcc      MVCCStateManager
	baseState StateDB
	txIndex   int
}

// NewStateWrapper creates a new state wrapper for a specific transaction
func NewStateWrapper(mvcc MVCCStateManager, baseState StateDB, txIndex int) StateDB {
	return &StateWrapper{
		mvcc:      mvcc,
		baseState: baseState,
		txIndex:   txIndex,
	}
}

// GetAccount retrieves an account, checking MVCC first, then base state
func (sw *StateWrapper) GetAccount(addr common.Address) (*accounts.Account, error) {
	// Try to read from MVCC
	account, err := sw.mvcc.ReadAccount(sw.txIndex, addr)
	if err != nil {
		return nil, err
	}

	// If found in MVCC, return it
	if account != nil {
		return account, nil
	}

	// Otherwise, read from base state
	return sw.baseState.GetAccount(addr)
}

// SetAccount writes an account to MVCC
func (sw *StateWrapper) SetAccount(addr common.Address, account *accounts.Account) error {
	sw.mvcc.WriteAccount(sw.txIndex, addr, account)
	return nil
}

// GetState retrieves a storage value, checking MVCC first, then base state
func (sw *StateWrapper) GetState(addr common.Address, key common.Hash) (common.Hash, error) {
	// Try to read from MVCC
	value, err := sw.mvcc.ReadStorage(sw.txIndex, addr, key)
	if err != nil {
		return common.Hash{}, err
	}

	// If found in MVCC, return it
	if value != nil && !value.IsZero() {
		return common.BytesToHash(value.Bytes()), nil
	}

	// Otherwise, read from base state
	return sw.baseState.GetState(addr, key)
}

// SetState writes a storage value to MVCC
func (sw *StateWrapper) SetState(addr common.Address, key, value common.Hash) error {
	valueUint := uint256.NewInt(0).SetBytes(value.Bytes())
	sw.mvcc.WriteStorage(sw.txIndex, addr, key, valueUint)
	return nil
}

// GetCode retrieves code, checking MVCC first, then base state
func (sw *StateWrapper) GetCode(addr common.Address) ([]byte, error) {
	// Try to read from MVCC
	code, err := sw.mvcc.ReadCode(sw.txIndex, addr)
	if err != nil {
		return nil, err
	}

	// If found in MVCC, return it
	if code != nil {
		return code, nil
	}

	// Otherwise, read from base state
	return sw.baseState.GetCode(addr)
}

// SetCode writes code to MVCC
func (sw *StateWrapper) SetCode(addr common.Address, code []byte) error {
	sw.mvcc.WriteCode(sw.txIndex, addr, code)
	return nil
}

// GetBalance retrieves balance, checking MVCC first, then base state
func (sw *StateWrapper) GetBalance(addr common.Address) (*uint256.Int, error) {
	// Try to read from MVCC
	balance, err := sw.mvcc.ReadBalance(sw.txIndex, addr)
	if err != nil {
		return nil, err
	}

	// If found in MVCC, return it
	if balance != nil && !balance.IsZero() {
		return balance, nil
	}

	// Otherwise, read from base state
	return sw.baseState.GetBalance(addr)
}

// SetBalance writes balance to MVCC
func (sw *StateWrapper) SetBalance(addr common.Address, balance *uint256.Int) error {
	sw.mvcc.WriteBalance(sw.txIndex, addr, balance)
	return nil
}

// GetNonce retrieves nonce, checking MVCC first, then base state
func (sw *StateWrapper) GetNonce(addr common.Address) (uint64, error) {
	// Try to read from MVCC
	nonce, err := sw.mvcc.ReadNonce(sw.txIndex, addr)
	if err != nil {
		return 0, err
	}

	// If found in MVCC (non-zero), return it
	if nonce > 0 {
		return nonce, nil
	}

	// Otherwise, read from base state
	return sw.baseState.GetNonce(addr)
}

// SetNonce writes nonce to MVCC
func (sw *StateWrapper) SetNonce(addr common.Address, nonce uint64) error {
	sw.mvcc.WriteNonce(sw.txIndex, addr, nonce)
	return nil
}

// GetReadSet returns the read set for this transaction
func (sw *StateWrapper) GetReadSet() *ReadSet {
	return sw.mvcc.GetReadSet(sw.txIndex)
}

// GetWriteSet returns the write set for this transaction
func (sw *StateWrapper) GetWriteSet() *WriteSet {
	return sw.mvcc.GetWriteSet(sw.txIndex)
}

// MVCCStateDB extends StateDB with MVCC-specific methods
type MVCCStateDB interface {
	StateDB
	GetReadSet() *ReadSet
	GetWriteSet() *WriteSet
}

// Ensure StateWrapper implements MVCCStateDB
var _ MVCCStateDB = (*StateWrapper)(nil)

// CreateStateWrappers creates state wrappers for all transactions
func CreateStateWrappers(
	mvcc MVCCStateManager,
	baseState StateDB,
	txCount int,
) []MVCCStateDB {
	wrappers := make([]MVCCStateDB, txCount)
	for i := 0; i < txCount; i++ {
		wrappers[i] = NewStateWrapper(mvcc, baseState, i).(MVCCStateDB)
	}
	return wrappers
}

// ValidateStateWrapper checks if a state wrapper is valid
func ValidateStateWrapper(sw StateDB) error {
	if sw == nil {
		return fmt.Errorf("state wrapper is nil")
	}
	return nil
}
