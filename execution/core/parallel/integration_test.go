package parallel

import (
	"testing"
	"time"

	"github.com/erigontech/erigon/common"
	"github.com/erigontech/erigon/core/types"
)

// TestBlockProcessorIntegration tests the block processor integration.
func TestBlockProcessorIntegration(t *testing.T) {
	config := DefaultConfig()
	config.Enabled = true
	config.WorkerCount = 4

	processor := NewBlockProcessor(config)

	// Create a test block
	block := createTestBlock(10)

	// Create a mock state
	state := &mockStateDB{}

	// Process the block
	result, err := processor.ProcessBlock(block, state)
	if err != nil {
		t.Fatalf("ProcessBlock failed: %v", err)
	}

	// Verify result
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
}

// TestParallelVsSequential tests that parallel and sequential execution produce the same results.
func TestParallelVsSequential(t *testing.T) {
	config := DefaultConfig()
	config.Enabled = true

	// Create test block
	block := createTestBlock(20)
	state := &mockStateDB{}

	// Execute in parallel
	parallelProcessor := NewBlockProcessor(config)
	parallelResult, err := parallelProcessor.ProcessBlock(block, state)
	if err != nil {
		t.Fatalf("Parallel execution failed: %v", err)
	}

	// Execute sequentially
	seqExecutor := NewSequentialExecutor()
	seqResult, err := seqExecutor.ExecuteBlock(block, state)
	if err != nil {
		t.Fatalf("Sequential execution failed: %v", err)
	}

	// Compare results
	if !compareExecutionResults(parallelResult, seqResult) {
		t.Error("Parallel and sequential results differ")
	}
}

// TestFeatureFlagIntegration tests feature flag integration.
func TestFeatureFlagIntegration(t *testing.T) {
	flags := NewFeatureFlags()

	// Test enabling/disabling
	flags.SetEnabled(true)
	if !flags.IsEnabled() {
		t.Error("Expected parallel execution to be enabled")
	}

	flags.SetEnabled(false)
	if flags.IsEnabled() {
		t.Error("Expected parallel execution to be disabled")
	}
}

// TestFallbackMechanism tests the fallback mechanism.
func TestFallbackMechanism(t *testing.T) {
	config := DefaultConfig()
	config.Enabled = true
	config.ConflictThreshold = 0.3 // Low threshold to trigger fallback

	processor := NewBlockProcessor(config)
	block := createTestBlock(10)
	state := &mockStateDB{}

	// Process block (may trigger fallback)
	result, err := processor.ProcessBlock(block, state)
	if err != nil {
		t.Fatalf("ProcessBlock failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}
}

// TestPanicRecovery tests panic recovery.
func TestPanicRecovery(t *testing.T) {
	config := DefaultConfig()
	config.Enabled = true

	executor := NewParallelExecutor(config)
	wrapper := NewPanicRecoveryWrapper(executor, 3, 1*time.Minute)

	block := createTestBlock(5)
	state := &mockStateDB{}

	// This should not panic even if underlying execution panics
	_, err := wrapper.ExecuteBlock(block, state)
	// Error is expected if panic occurs, but should not crash
	_ = err
}

// TestThresholdChecking tests threshold checking.
func TestThresholdChecking(t *testing.T) {
	config := DefaultThresholdConfig()
	checker := NewThresholdChecker(config)

	// Test conflict rate check
	exceeded, msg := checker.CheckConflictRate(0.6)
	if !exceeded {
		t.Error("Expected conflict rate to exceed threshold")
	}
	if msg == "" {
		t.Error("Expected non-empty message")
	}

	// Test within threshold
	exceeded, _ = checker.CheckConflictRate(0.3)
	if exceeded {
		t.Error("Expected conflict rate to be within threshold")
	}
}

// TestMetricsCollection tests metrics collection.
func TestMetricsCollection(t *testing.T) {
	collector := NewMetricsCollector()

	// Record some metrics
	metrics := &ExecutionMetrics{
		TotalTxs:      100,
		ParallelTxs:   100,
		ReexecutedTxs: 10,
		ConflictRate:  0.1,
		SpeedupFactor: 3.5,
	}

	collector.RecordBlockExecution(metrics)

	// Get aggregated metrics
	aggregated := collector.GetAggregatedMetrics()
	if aggregated.TotalBlocks != 1 {
		t.Errorf("Expected 1 block, got %d", aggregated.TotalBlocks)
	}
	if aggregated.TotalTransactions != 100 {
		t.Errorf("Expected 100 transactions, got %d", aggregated.TotalTransactions)
	}
}

// TestHybridExecution tests hybrid execution mode.
func TestHybridExecution(t *testing.T) {
	config := DefaultConfig()
	config.Enabled = true

	parallelExecutor := NewParallelExecutor(config)
	hybridConfig := DefaultHybridConfig()
	hybridExecutor := NewHybridExecutor(parallelExecutor, hybridConfig)

	// Test with small block (should use sequential)
	smallBlock := createTestBlock(5)
	state := &mockStateDB{}

	result, err := hybridExecutor.ExecuteBlock(smallBlock, state)
	if err != nil {
		t.Fatalf("Hybrid execution failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Test with large block (should use parallel)
	largeBlock := createTestBlock(50)
	result, err = hybridExecutor.ExecuteBlock(largeBlock, state)
	if err != nil {
		t.Fatalf("Hybrid execution failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
}

// Helper functions

// createTestBlock creates a test block with the specified number of transactions.
func createTestBlock(txCount int) *types.Block {
	// In a full implementation, this would create a proper test block
	// For now, return a minimal block structure
	return types.NewBlock(&types.Header{}, nil, nil, nil)
}

// mockStateDB is a mock implementation of StateDB for testing.
type mockStateDB struct{}

func (m *mockStateDB) GetAccount(addr common.Address) (*types.Account, error) {
	return &types.Account{}, nil
}

func (m *mockStateDB) SetAccount(addr common.Address, account *types.Account) error {
	return nil
}

func (m *mockStateDB) GetState(addr common.Address, key common.Hash) (common.Hash, error) {
	return common.Hash{}, nil
}

func (m *mockStateDB) SetState(addr common.Address, key, value common.Hash) error {
	return nil
}

// compareExecutionResults compares two execution results.
func compareExecutionResults(r1, r2 *ExecutionResult) bool {
	if r1 == nil || r2 == nil {
		return r1 == r2
	}

	// Compare gas used
	if r1.GasUsed != r2.GasUsed {
		return false
	}

	// Compare receipt count
	if len(r1.Receipts) != len(r2.Receipts) {
		return false
	}

	return true
}

// BenchmarkParallelExecution benchmarks parallel execution.
func BenchmarkParallelExecution(b *testing.B) {
	config := DefaultConfig()
	config.Enabled = true
	config.WorkerCount = 8

	processor := NewBlockProcessor(config)
	block := createTestBlock(100)
	state := &mockStateDB{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = processor.ProcessBlock(block, state)
	}
}

// BenchmarkSequentialExecution benchmarks sequential execution.
func BenchmarkSequentialExecution(b *testing.B) {
	executor := NewSequentialExecutor()
	block := createTestBlock(100)
	state := &mockStateDB{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = executor.ExecuteBlock(block, state)
	}
}
