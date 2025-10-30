package mvcc

import (
	"testing"

	"github.com/erigontech/erigon/common"
	"github.com/erigontech/erigon/core/types"
)

// TestMVCCStateManagerReadWrite tests basic read/write operations
func TestMVCCStateManagerReadWrite(t *testing.T) {
	manager := NewManager(nil)

	addr := common.HexToAddress("0x1234")
	account := &types.Account{Nonce: 42}

	// Write account for tx 0
	manager.WriteAccount(0, addr, account)

	// Read from tx 1 (should see tx 0's write)
	readAccount, err := manager.ReadAccount(1, addr)
	if err != nil {
		t.Fatalf("ReadAccount failed: %v", err)
	}
	if readAccount.Nonce != 42 {
		t.Errorf("Expected nonce 42, got %d", readAccount.Nonce)
	}
}

// TestMVCCVersionVisibility tests that transactions see correct versions
func TestMVCCVersionVisibility(t *testing.T) {
	manager := NewManager(nil)

	addr := common.HexToAddress("0x1234")

	// Tx 0 writes nonce 1
	manager.WriteAccount(0, addr, &types.Account{Nonce: 1})

	// Tx 2 writes nonce 2
	manager.WriteAccount(2, addr, &types.Account{Nonce: 2})

	// Tx 1 should see tx 0's write (nonce 1)
	account1, _ := manager.ReadAccount(1, addr)
	if account1.Nonce != 1 {
		t.Errorf("Tx 1 should see nonce 1, got %d", account1.Nonce)
	}

	// Tx 3 should see tx 2's write (nonce 2)
	account3, _ := manager.ReadAccount(3, addr)
	if account3.Nonce != 2 {
		t.Errorf("Tx 3 should see nonce 2, got %d", account3.Nonce)
	}
}

// TestReadSetTracking tests that read sets are tracked correctly
func TestReadSetTracking(t *testing.T) {
	manager := NewManager(nil)

	addr1 := common.HexToAddress("0x1111")
	addr2 := common.HexToAddress("0x2222")

	// Tx 0 reads two accounts
	manager.ReadAccount(0, addr1)
	manager.ReadAccount(0, addr2)

	// Get read set
	readSet := manager.GetReadSet(0)
	if readSet == nil {
		t.Fatal("Expected non-nil read set")
	}

	if !readSet.Accounts[addr1] {
		t.Error("Expected addr1 in read set")
	}
	if !readSet.Accounts[addr2] {
		t.Error("Expected addr2 in read set")
	}
}

// TestWriteSetTracking tests that write sets are tracked correctly
func TestWriteSetTracking(t *testing.T) {
	manager := NewManager(nil)

	addr := common.HexToAddress("0x1234")
	account := &types.Account{Nonce: 42}

	// Tx 0 writes account
	manager.WriteAccount(0, addr, account)

	// Get write set
	writeSet := manager.GetWriteSet(0)
	if writeSet == nil {
		t.Fatal("Expected non-nil write set")
	}

	if _, exists := writeSet.Accounts[addr]; !exists {
		t.Error("Expected addr in write set")
	}
}

// TestStorageReadWrite tests storage operations
func TestStorageReadWrite(t *testing.T) {
	manager := NewManager(nil)

	addr := common.HexToAddress("0x1234")
	key := common.HexToHash("0xabcd")
	value := common.HexToHash("0x5678")

	// Tx 0 writes storage
	manager.WriteStorage(0, addr, key, value)

	// Tx 1 reads storage
	readValue, err := manager.ReadStorage(1, addr, key)
	if err != nil {
		t.Fatalf("ReadStorage failed: %v", err)
	}
	if readValue != value {
		t.Errorf("Expected value %v, got %v", value, readValue)
	}
}

// TestRemoveTransaction tests removing a transaction's versions
func TestRemoveTransaction(t *testing.T) {
	manager := NewManager(nil)

	addr := common.HexToAddress("0x1234")

	// Tx 0 writes
	manager.WriteAccount(0, addr, &types.Account{Nonce: 1})

	// Tx 1 writes
	manager.WriteAccount(1, addr, &types.Account{Nonce: 2})

	// Remove tx 1
	manager.RemoveTransaction(1)

	// Tx 2 should now see tx 0's write (nonce 1)
	account, _ := manager.ReadAccount(2, addr)
	if account.Nonce != 1 {
		t.Errorf("After removing tx 1, should see nonce 1, got %d", account.Nonce)
	}
}

// TestReset tests resetting the manager
func TestReset(t *testing.T) {
	manager := NewManager(nil)

	addr := common.HexToAddress("0x1234")
	manager.WriteAccount(0, addr, &types.Account{Nonce: 42})

	// Reset
	manager.Reset()

	// Read set should be empty
	readSet := manager.GetReadSet(0)
	if readSet != nil && len(readSet.Accounts) > 0 {
		t.Error("Expected empty read set after reset")
	}
}

// TestHasReadWriteConflict tests conflict detection
func TestHasReadWriteConflict(t *testing.T) {
	readSet := &ReadSet{
		Accounts: make(map[common.Address]bool),
		Storage:  make(map[StorageKey]bool),
	}
	writeSet := &WriteSet{
		Accounts: make(map[common.Address]*types.Account),
		Storage:  make(map[StorageKey]common.Hash),
	}

	addr := common.HexToAddress("0x1234")

	// No conflict initially
	if HasReadWriteConflict(readSet, writeSet) {
		t.Error("Should not have conflict with empty sets")
	}

	// Add same address to both
	readSet.Accounts[addr] = true
	writeSet.Accounts[addr] = &types.Account{Nonce: 1}

	// Should have conflict now
	if !HasReadWriteConflict(readSet, writeSet) {
		t.Error("Should have conflict when same address in both sets")
	}
}

// TestConcurrentAccess tests thread-safe concurrent access
func TestConcurrentAccess(t *testing.T) {
	manager := NewManager(nil)

	addr := common.HexToAddress("0x1234")

	// Concurrent writes
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(txIndex int) {
			manager.WriteAccount(txIndex, addr, &types.Account{Nonce: uint64(txIndex)})
			done <- true
		}(i)
	}

	// Wait for all writes
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic and should have versions
	account, _ := manager.ReadAccount(100, addr)
	if account == nil {
		t.Error("Expected to read an account")
	}
}

// BenchmarkMVCCWrite benchmarks write operations
func BenchmarkMVCCWrite(b *testing.B) {
	manager := NewManager(nil)
	addr := common.HexToAddress("0x1234")
	account := &types.Account{Nonce: 42}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.WriteAccount(i, addr, account)
	}
}

// BenchmarkMVCCRead benchmarks read operations
func BenchmarkMVCCRead(b *testing.B) {
	manager := NewManager(nil)
	addr := common.HexToAddress("0x1234")

	// Pre-populate with some versions
	for i := 0; i < 100; i++ {
		manager.WriteAccount(i, addr, &types.Account{Nonce: uint64(i)})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.ReadAccount(i%100, addr)
	}
}
