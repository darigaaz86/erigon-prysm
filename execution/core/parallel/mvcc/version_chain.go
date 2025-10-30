package mvcc

import (
	"sync"
)

// VersionChain manages a linked list of versions for a single state variable.
// Each version is associated with a transaction index.
type VersionChain struct {
	mu   sync.RWMutex
	head *VersionNode
}

// VersionNode represents a single version in the chain
type VersionNode struct {
	TxIndex int
	Value   interface{}
	Next    *VersionNode
}

// NewVersionChain creates a new version chain
func NewVersionChain() *VersionChain {
	return &VersionChain{
		head: nil,
	}
}

// Write adds a new version to the chain
func (vc *VersionChain) Write(txIndex int, value interface{}) {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	// Create new node
	node := &VersionNode{
		TxIndex: txIndex,
		Value:   value,
		Next:    vc.head,
	}

	// Insert at head (most recent version)
	vc.head = node
}

// Read returns the latest version visible to the given transaction
// A version is visible if its txIndex < the reading transaction's txIndex
func (vc *VersionChain) Read(txIndex int) interface{} {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	// Traverse the chain to find the latest visible version
	current := vc.head
	for current != nil {
		if current.TxIndex < txIndex {
			return current.Value
		}
		current = current.Next
	}

	// No visible version found
	return nil
}

// GetAllVersions returns all versions in the chain (for debugging)
func (vc *VersionChain) GetAllVersions() []*VersionNode {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	var versions []*VersionNode
	current := vc.head
	for current != nil {
		versions = append(versions, current)
		current = current.Next
	}
	return versions
}

// Clear removes all versions from the chain
func (vc *VersionChain) Clear() {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	vc.head = nil
}

// HasVersion checks if a version exists for the given transaction
func (vc *VersionChain) HasVersion(txIndex int) bool {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	current := vc.head
	for current != nil {
		if current.TxIndex == txIndex {
			return true
		}
		current = current.Next
	}
	return false
}

// GetVersion returns the version for a specific transaction index
func (vc *VersionChain) GetVersion(txIndex int) interface{} {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	current := vc.head
	for current != nil {
		if current.TxIndex == txIndex {
			return current.Value
		}
		current = current.Next
	}
	return nil
}

// Length returns the number of versions in the chain
func (vc *VersionChain) Length() int {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	count := 0
	current := vc.head
	for current != nil {
		count++
		current = current.Next
	}
	return count
}
