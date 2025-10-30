package mvcc

import (
	"sync"

	"github.com/erigontech/erigon/common"
	"github.com/erigontech/erigon/execution/types/accounts"
	"github.com/holiman/uint256"
)

// MVCCStateManager manages multiple versions of state variables for parallel execution
// using Erigon's production types.
type MVCCStateManager interface {
	// Account operations
	ReadAccount(txIndex int, addr common.Address) (*accounts.Account, error)
	WriteAccount(txIndex int, addr common.Address, account *accounts.Account)

	// Storage operations
	ReadStorage(txIndex int, addr common.Address, key common.Hash) (*uint256.Int, error)
	WriteStorage(txIndex int, addr common.Address, key common.Hash, value *uint256.Int)

	// Code operations
	ReadCode(txIndex int, addr common.Address) ([]byte, error)
	WriteCode(txIndex int, addr common.Address, code []byte)

	// Balance operations (convenience)
	ReadBalance(txIndex int, addr common.Address) (*uint256.Int, error)
	WriteBalance(txIndex int, addr common.Address, balance *uint256.Int)

	// Nonce operations (convenience)
	ReadNonce(txIndex int, addr common.Address) (uint64, error)
	WriteNonce(txIndex int, addr common.Address, nonce uint64)

	// Read/Write set tracking
	GetReadSet(txIndex int) *ReadSet
	GetWriteSet(txIndex int) *WriteSet

	// Reset for new block
	Reset()
}

// manager implements MVCCStateManager
type manager struct {
	mu sync.RWMutex

	// Version chains for different state types
	accounts map[common.Address]*VersionChain
	storage  map[StorageKey]*VersionChain
	code     map[common.Address]*VersionChain

	// Read and write sets for each transaction
	readSets  map[int]*ReadSet
	writeSets map[int]*WriteSet
}

// NewManager creates a new MVCC state manager
func NewManager(baseState interface{}) MVCCStateManager {
	return &manager{
		accounts:  make(map[common.Address]*VersionChain),
		storage:   make(map[StorageKey]*VersionChain),
		code:      make(map[common.Address]*VersionChain),
		readSets:  make(map[int]*ReadSet),
		writeSets: make(map[int]*WriteSet),
	}
}

// ReadAccount returns the account visible to the given transaction
func (m *manager) ReadAccount(txIndex int, addr common.Address) (*accounts.Account, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get version chain for this address
	chain, exists := m.accounts[addr]
	if !exists {
		// No versions yet, return nil (will read from base state)
		m.trackAccountRead(txIndex, addr)
		return nil, nil
	}

	// Find the latest version visible to this transaction
	value := chain.Read(txIndex)
	m.trackAccountRead(txIndex, addr)

	if value == nil {
		return nil, nil
	}

	return value.(*accounts.Account), nil
}

// WriteAccount records a new version of an account
func (m *manager) WriteAccount(txIndex int, addr common.Address, account *accounts.Account) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get or create version chain
	chain, exists := m.accounts[addr]
	if !exists {
		chain = NewVersionChain()
		m.accounts[addr] = chain
	}

	// Add new version
	chain.Write(txIndex, account)
	m.trackAccountWrite(txIndex, addr, account)
}

// ReadStorage returns storage value visible to the given transaction
func (m *manager) ReadStorage(txIndex int, addr common.Address, key common.Hash) (*uint256.Int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	storageKey := StorageKey{Address: addr, Slot: key}

	// Get version chain for this storage slot
	chain, exists := m.storage[storageKey]
	if !exists {
		// No versions yet, return zero
		m.trackStorageRead(txIndex, addr, key)
		return uint256.NewInt(0), nil
	}

	// Find the latest version visible to this transaction
	value := chain.Read(txIndex)
	m.trackStorageRead(txIndex, addr, key)

	if value == nil {
		return uint256.NewInt(0), nil
	}

	return value.(*uint256.Int), nil
}

// WriteStorage records a new version of a storage slot
func (m *manager) WriteStorage(txIndex int, addr common.Address, key common.Hash, value *uint256.Int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	storageKey := StorageKey{Address: addr, Slot: key}

	// Get or create version chain
	chain, exists := m.storage[storageKey]
	if !exists {
		chain = NewVersionChain()
		m.storage[storageKey] = chain
	}

	// Add new version
	chain.Write(txIndex, value)
	m.trackStorageWrite(txIndex, addr, key, value)
}

// ReadCode returns code visible to the given transaction
func (m *manager) ReadCode(txIndex int, addr common.Address) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get version chain for this address
	chain, exists := m.code[addr]
	if !exists {
		m.trackCodeRead(txIndex, addr)
		return nil, nil
	}

	// Find the latest version visible to this transaction
	value := chain.Read(txIndex)
	m.trackCodeRead(txIndex, addr)

	if value == nil {
		return nil, nil
	}

	return value.([]byte), nil
}

// WriteCode records a new version of code
func (m *manager) WriteCode(txIndex int, addr common.Address, code []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get or create version chain
	chain, exists := m.code[addr]
	if !exists {
		chain = NewVersionChain()
		m.code[addr] = chain
	}

	// Add new version
	chain.Write(txIndex, code)
	m.trackCodeWrite(txIndex, addr, code)
}

// ReadBalance is a convenience method to read just the balance
func (m *manager) ReadBalance(txIndex int, addr common.Address) (*uint256.Int, error) {
	account, err := m.ReadAccount(txIndex, addr)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return uint256.NewInt(0), nil
	}
	return &account.Balance, nil
}

// WriteBalance is a convenience method to write just the balance
func (m *manager) WriteBalance(txIndex int, addr common.Address, balance *uint256.Int) {
	// Read current account or create new one
	account, _ := m.ReadAccount(txIndex, addr)
	if account == nil {
		account = &accounts.Account{
			Balance: *balance,
		}
	} else {
		account.Balance = *balance
	}
	m.WriteAccount(txIndex, addr, account)
}

// ReadNonce is a convenience method to read just the nonce
func (m *manager) ReadNonce(txIndex int, addr common.Address) (uint64, error) {
	account, err := m.ReadAccount(txIndex, addr)
	if err != nil {
		return 0, err
	}
	if account == nil {
		return 0, nil
	}
	return account.Nonce, nil
}

// WriteNonce is a convenience method to write just the nonce
func (m *manager) WriteNonce(txIndex int, addr common.Address, nonce uint64) {
	// Read current account or create new one
	account, _ := m.ReadAccount(txIndex, addr)
	if account == nil {
		account = &accounts.Account{
			Nonce: nonce,
		}
	} else {
		account.Nonce = nonce
	}
	m.WriteAccount(txIndex, addr, account)
}

// GetReadSet returns all state variables read by a transaction
func (m *manager) GetReadSet(txIndex int) *ReadSet {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rs, exists := m.readSets[txIndex]
	if !exists {
		return NewReadSet()
	}
	return rs
}

// GetWriteSet returns all state variables written by a transaction
func (m *manager) GetWriteSet(txIndex int) *WriteSet {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ws, exists := m.writeSets[txIndex]
	if !exists {
		return NewWriteSet()
	}
	return ws
}

// Reset clears all versions for a new block
func (m *manager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.accounts = make(map[common.Address]*VersionChain)
	m.storage = make(map[StorageKey]*VersionChain)
	m.code = make(map[common.Address]*VersionChain)
	m.readSets = make(map[int]*ReadSet)
	m.writeSets = make(map[int]*WriteSet)
}

// Helper methods for tracking reads and writes
// NOTE: These methods assume the caller already holds the appropriate lock

func (m *manager) trackAccountRead(txIndex int, addr common.Address) {
	rs, exists := m.readSets[txIndex]
	if !exists {
		rs = NewReadSet()
		m.readSets[txIndex] = rs
	}
	rs.Accounts[addr] = true
}

func (m *manager) trackAccountWrite(txIndex int, addr common.Address, account *accounts.Account) {
	ws, exists := m.writeSets[txIndex]
	if !exists {
		ws = NewWriteSet()
		m.writeSets[txIndex] = ws
	}
	ws.Accounts[addr] = account
}

func (m *manager) trackStorageRead(txIndex int, addr common.Address, key common.Hash) {
	rs, exists := m.readSets[txIndex]
	if !exists {
		rs = NewReadSet()
		m.readSets[txIndex] = rs
	}
	storageKey := StorageKey{Address: addr, Slot: key}
	rs.Storage[storageKey] = true
}

func (m *manager) trackStorageWrite(txIndex int, addr common.Address, key common.Hash, value *uint256.Int) {
	ws, exists := m.writeSets[txIndex]
	if !exists {
		ws = NewWriteSet()
		m.writeSets[txIndex] = ws
	}
	storageKey := StorageKey{Address: addr, Slot: key}
	ws.Storage[storageKey] = value
}

func (m *manager) trackCodeRead(txIndex int, addr common.Address) {
	rs, exists := m.readSets[txIndex]
	if !exists {
		rs = NewReadSet()
		m.readSets[txIndex] = rs
	}
	rs.Code[addr] = true
}

func (m *manager) trackCodeWrite(txIndex int, addr common.Address, code []byte) {
	ws, exists := m.writeSets[txIndex]
	if !exists {
		ws = NewWriteSet()
		m.writeSets[txIndex] = ws
	}
	ws.Code[addr] = code
}

// StorageKey uniquely identifies a storage slot
type StorageKey struct {
	Address common.Address
	Slot    common.Hash
}

// ReadSet tracks all state variables read by a transaction
type ReadSet struct {
	Accounts map[common.Address]bool
	Storage  map[StorageKey]bool
	Code     map[common.Address]bool
	Balances map[common.Address]bool
	Nonces   map[common.Address]bool
}

// NewReadSet creates a new ReadSet
func NewReadSet() *ReadSet {
	return &ReadSet{
		Accounts: make(map[common.Address]bool),
		Storage:  make(map[StorageKey]bool),
		Code:     make(map[common.Address]bool),
		Balances: make(map[common.Address]bool),
		Nonces:   make(map[common.Address]bool),
	}
}

// WriteSet tracks all state variables written by a transaction
type WriteSet struct {
	Accounts map[common.Address]*accounts.Account
	Storage  map[StorageKey]*uint256.Int
	Code     map[common.Address][]byte
	Balances map[common.Address]*uint256.Int
	Nonces   map[common.Address]uint64
}

// NewWriteSet creates a new WriteSet
func NewWriteSet() *WriteSet {
	return &WriteSet{
		Accounts: make(map[common.Address]*accounts.Account),
		Storage:  make(map[StorageKey]*uint256.Int),
		Code:     make(map[common.Address][]byte),
		Balances: make(map[common.Address]*uint256.Int),
		Nonces:   make(map[common.Address]uint64),
	}
}
