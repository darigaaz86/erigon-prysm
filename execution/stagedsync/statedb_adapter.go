package stagedsync

import (
	"fmt"

	"github.com/erigontech/erigon/common"
	"github.com/erigontech/erigon/db/kv"
	dbstate "github.com/erigontech/erigon/db/state"
	"github.com/erigontech/erigon/execution/core/parallel"
	"github.com/erigontech/erigon/execution/state"
	"github.com/erigontech/erigon/execution/types/accounts"
	"github.com/holiman/uint256"
)

// StateDBAdapter adapts SharedDomains to implement parallel.StateDB interface
type StateDBAdapter struct {
	domains  *dbstate.SharedDomains
	tx       kv.TemporalTx
	reader   state.StateReader
	blockNum uint64
	txNum    uint64
}

// NewStateDBAdapter creates a new StateDB adapter
func NewStateDBAdapter(domains *dbstate.SharedDomains, tx kv.TemporalTx, blockNum, txNum uint64) parallel.StateDB {
	reader := state.NewReaderV3(domains.AsGetter(tx))
	return &StateDBAdapter{
		domains:  domains,
		tx:       tx,
		reader:   reader,
		blockNum: blockNum,
		txNum:    txNum,
	}
}

// GetAccount retrieves an account from state
func (s *StateDBAdapter) GetAccount(addr common.Address) (*accounts.Account, error) {
	// Read account from domains
	acc, err := s.reader.ReadAccountData(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to read account: %w", err)
	}

	if acc == nil {
		// Account doesn't exist, return empty account
		acc = &accounts.Account{
			Nonce:   0,
			Balance: *uint256.NewInt(0),
		}
	}

	return acc, nil
}

// SetAccount updates an account in state
func (s *StateDBAdapter) SetAccount(addr common.Address, account *accounts.Account) error {
	// Write account to domains
	writer := state.NewWriter(s.domains.AsPutDel(s.tx), nil, s.txNum)
	if err := writer.UpdateAccountData(addr, account, nil); err != nil {
		return fmt.Errorf("failed to write account: %w", err)
	}

	return nil
}

// GetState retrieves a storage value
func (s *StateDBAdapter) GetState(addr common.Address, key common.Hash) (common.Hash, error) {
	// Read storage from domains
	value, _, err := s.reader.ReadAccountStorage(addr, key)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to read storage: %w", err)
	}

	// Convert uint256.Int to common.Hash
	return common.BytesToHash(value.Bytes()), nil
}

// SetState updates a storage value
func (s *StateDBAdapter) SetState(addr common.Address, key, value common.Hash) error {
	// Convert value to uint256
	valueUint256 := bigIntToUint256(value.Big())

	// Read original value
	original, _, err := s.reader.ReadAccountStorage(addr, key)
	if err != nil {
		// If read fails, assume original is zero
		original = *uint256.NewInt(0)
	}

	// Write storage to domains
	writer := state.NewWriter(s.domains.AsPutDel(s.tx), nil, s.txNum)
	if err := writer.WriteAccountStorage(addr, 0, key, original, *valueUint256); err != nil {
		return fmt.Errorf("failed to write storage: %w", err)
	}

	return nil
}


// GetCode retrieves code from state
func (s *StateDBAdapter) GetCode(addr common.Address) ([]byte, error) {
	code, err := s.reader.ReadAccountCode(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to read code: %w", err)
	}
	return code, nil
}

// SetCode updates code in state
func (s *StateDBAdapter) SetCode(addr common.Address, code []byte) error {
	writer := state.NewWriter(s.domains.AsPutDel(s.tx), nil, s.txNum)
	if err := writer.UpdateAccountCode(addr, 0, common.Hash{}, code); err != nil {
		return fmt.Errorf("failed to write code: %w", err)
	}
	return nil
}

// GetBalance retrieves balance from state
func (s *StateDBAdapter) GetBalance(addr common.Address) (*uint256.Int, error) {
	acc, err := s.GetAccount(addr)
	if err != nil {
		return nil, err
	}
	return &acc.Balance, nil
}

// SetBalance updates balance in state
func (s *StateDBAdapter) SetBalance(addr common.Address, balance *uint256.Int) error {
	acc, err := s.GetAccount(addr)
	if err != nil {
		return err
	}
	acc.Balance = *balance
	return s.SetAccount(addr, acc)
}

// GetNonce retrieves nonce from state
func (s *StateDBAdapter) GetNonce(addr common.Address) (uint64, error) {
	acc, err := s.GetAccount(addr)
	if err != nil {
		return 0, err
	}
	return acc.Nonce, nil
}

// SetNonce updates nonce in state
func (s *StateDBAdapter) SetNonce(addr common.Address, nonce uint64) error {
	acc, err := s.GetAccount(addr)
	if err != nil {
		return err
	}
	acc.Nonce = nonce
	return s.SetAccount(addr, acc)
}
