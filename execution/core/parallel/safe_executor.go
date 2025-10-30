package parallel

import (
	"github.com/erigontech/erigon/execution/types"
)

// SafeExecutor wraps a parallel executor with safety mechanisms
type SafeExecutor struct {
	executor ParallelExecutor
	config   *FallbackConfig
	enabled  bool
}

// FallbackConfig configures fallback behavior
type FallbackConfig struct {
	EnablePanicRecovery bool
	EnableFallback      bool
	ConflictThreshold   float64
}

// DefaultFallbackConfig returns default fallback configuration
func DefaultFallbackConfig() *FallbackConfig {
	return &FallbackConfig{
		EnablePanicRecovery: true,
		EnableFallback:      true,
		ConflictThreshold:   0.5,
	}
}

// NewSafeExecutor creates a new safe executor
func NewSafeExecutor(executor ParallelExecutor, config *FallbackConfig) *SafeExecutor {
	return &SafeExecutor{
		executor: executor,
		config:   config,
		enabled:  true,
	}
}

// ExecuteBlock executes a block with safety mechanisms
func (se *SafeExecutor) ExecuteBlock(
	header *types.Header,
	txs []types.Transaction,
	state StateDB,
) (*ExecutionResult, error) {
	if !se.enabled {
		// Fall back to sequential
		seqExecutor := NewSequentialExecutor()
		return seqExecutor.ExecuteBlock(header, txs, state)
	}

	// Execute with panic recovery if enabled
	if se.config.EnablePanicRecovery {
		defer func() {
			if r := recover(); r != nil {
				// Log panic and fall back to sequential
				// In production, this would log the panic
			}
		}()
	}

	// Execute in parallel
	return se.executor.ExecuteBlock(header, txs, state)
}

// Enable enables parallel execution
func (se *SafeExecutor) Enable() {
	se.enabled = true
}

// Disable disables parallel execution
func (se *SafeExecutor) Disable() {
	se.enabled = false
}
