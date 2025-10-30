package parallel

import (
	"fmt"

	"github.com/erigontech/erigon/execution/types"
)

// BlockProcessor integrates parallel execution with Erigon's block processing using production types.
type BlockProcessor struct {
	parallelExecutor ParallelExecutor
	safeExecutor     *SafeExecutor
	config           *Config
	enabled          bool
}

// NewBlockProcessor creates a new block processor with parallel execution support.
func NewBlockProcessor(config *Config) *BlockProcessor {
	if config == nil {
		config = DefaultConfig()
	}

	parallelExecutor := NewParallelExecutor(config)
	safeExecutor := NewSafeExecutor(parallelExecutor, DefaultFallbackConfig())

	return &BlockProcessor{
		parallelExecutor: parallelExecutor,
		safeExecutor:     safeExecutor,
		config:           config,
		enabled:          config.Enabled,
	}
}

// ProcessBlock processes a block using parallel or sequential execution.
// Takes header and transactions separately instead of a Block object.
func (bp *BlockProcessor) ProcessBlock(
	header *types.Header,
	txs []types.Transaction,
	state StateDB,
) (*ExecutionResult, error) {
	// Check if parallel execution is enabled
	if !bp.enabled || !bp.config.Enabled {
		return bp.processSequential(header, txs, state)
	}

	// Check if block has enough transactions for parallel execution
	if len(txs) < 10 {
		return bp.processSequential(header, txs, state)
	}

	// Use safe executor with fallback mechanisms
	return bp.safeExecutor.ExecuteBlock(header, txs, state)
}

// processSequential processes a block sequentially (fallback mode).
func (bp *BlockProcessor) processSequential(
	header *types.Header,
	txs []types.Transaction,
	state StateDB,
) (*ExecutionResult, error) {
	seqExecutor := NewSequentialExecutor()
	return seqExecutor.ExecuteBlock(header, txs, state)
}

// Enable enables parallel execution.
func (bp *BlockProcessor) Enable() {
	bp.enabled = true
	bp.safeExecutor.Enable()
}

// Disable disables parallel execution.
func (bp *BlockProcessor) Disable() {
	bp.enabled = false
	bp.safeExecutor.Disable()
}

// IsEnabled returns whether parallel execution is enabled.
func (bp *BlockProcessor) IsEnabled() bool {
	return bp.enabled
}

// GetMetrics returns execution metrics.
func (bp *BlockProcessor) GetMetrics() *ExecutionMetrics {
	return bp.parallelExecutor.GetMetrics()
}

// IntegrationPoint provides hooks for integrating with Erigon's state processor.
type IntegrationPoint struct {
	processor *BlockProcessor
}

// NewIntegrationPoint creates a new integration point.
func NewIntegrationPoint(config *Config) *IntegrationPoint {
	return &IntegrationPoint{
		processor: NewBlockProcessor(config),
	}
}

// ExecuteBlock is the main entry point for block execution from Erigon.
func (ip *IntegrationPoint) ExecuteBlock(
	header *types.Header,
	txs []types.Transaction,
	state StateDB,
) (*ExecutionResult, error) {
	return ip.processor.ProcessBlock(header, txs, state)
}

// ShouldUseParallel determines if parallel execution should be used for a block.
func ShouldUseParallel(txCount int, config *Config) bool {
	// Don't use parallel for empty blocks
	if txCount == 0 {
		return false
	}

	// Don't use parallel if disabled
	if !config.Enabled {
		return false
	}

	// Don't use parallel for small transaction counts
	if txCount < 10 {
		return false
	}

	return true
}

// InitializeParallelExecution initializes parallel execution for Erigon.
func InitializeParallelExecution(config *Config) (*BlockProcessor, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create block processor
	processor := NewBlockProcessor(config)

	return processor, nil
}

// ShutdownParallelExecution gracefully shuts down parallel execution.
func ShutdownParallelExecution(processor *BlockProcessor) error {
	if processor == nil {
		return nil
	}

	// Disable parallel execution
	processor.Disable()

	return nil
}
