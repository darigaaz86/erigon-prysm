package stagedsync

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/erigontech/erigon/common/log/v3"
	"github.com/erigontech/erigon/db/kv"
	dbstate "github.com/erigontech/erigon/db/state"
	"github.com/erigontech/erigon/execution/core/parallel"
	"github.com/erigontech/erigon/execution/exec"
)

// ParallelExecutionAdapter adapts the parallel execution package to work with stagedsync
type ParallelExecutionAdapter struct {
	executor      *parallel.BlockProcessor
	reconstructor *BlockReconstructor
	logger        log.Logger
	enabled       bool
	workerCount   int
	enableMetrics bool
}

// NewParallelExecutionAdapter creates a new adapter for parallel execution
func NewParallelExecutionAdapter(logger log.Logger) *ParallelExecutionAdapter {
	adapter := &ParallelExecutionAdapter{
		logger:        logger,
		enabled:       false,
		reconstructor: NewBlockReconstructor(),
	}

	// Check if parallel execution is enabled via environment variable
	if enabled := os.Getenv("PARALLEL_ENABLED"); enabled == "true" || enabled == "1" {
		adapter.enabled = true

		// Create configuration from environment variables
		config := parallel.DefaultConfig()
		config.Enabled = true

		if workers := os.Getenv("PARALLEL_WORKERS"); workers != "" {
			if w, err := strconv.Atoi(workers); err == nil && w > 0 {
				config.WorkerCount = w
				adapter.workerCount = w
			}
		} else {
			adapter.workerCount = config.WorkerCount
		}

		if metrics := os.Getenv("PARALLEL_METRICS"); metrics == "true" || metrics == "1" {
			config.EnableMetrics = true
			adapter.enableMetrics = true
		}

		// Create the parallel executor
		adapter.executor = parallel.NewBlockProcessor(config)

		logger.Info("[Parallel Execution] Initialized",
			"workers", adapter.workerCount,
			"metrics", adapter.enableMetrics)
	}

	return adapter
}

// IsEnabled returns whether parallel execution is enabled
func (a *ParallelExecutionAdapter) IsEnabled() bool {
	return a.enabled && a.executor != nil
}

// ShouldUseParallel determines if a block should be executed in parallel
func (a *ParallelExecutionAdapter) ShouldUseParallel(txCount int) bool {
	if !a.IsEnabled() {
		// a.logger.Debug("[Parallel Execution] Not enabled")
		return false
	}

	// Only use parallel execution for blocks with 10+ transactions
	shouldUse := txCount >= 10
	// if shouldUse {
	// 	a.logger.Info("[Parallel Execution] Block qualifies for parallel execution", "txCount", txCount)
	// } else {
	// 	a.logger.Debug("[Parallel Execution] Block too small for parallel execution", "txCount", txCount, "threshold", 10)
	// }
	return shouldUse
}

// ExecuteBlock executes block tasks using parallel execution
func (a *ParallelExecutionAdapter) ExecuteBlock(
	ctx context.Context,
	tasks []exec.Task,
	domains *dbstate.SharedDomains,
	tx kv.TemporalTx,
	blockNum, txNum uint64,
) (*parallel.ExecutionResult, error) {
	if !a.IsEnabled() {
		return nil, fmt.Errorf("parallel execution is not enabled")
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("no tasks to execute")
	}

	// Count actual transactions
	txCount := 0
	for _, task := range tasks {
		txTask := task.(*exec.TxTask)
		if txTask.TxIndex >= 0 && !txTask.IsBlockEnd() {
			txCount++
		}
	}

	a.logger.Info("[Parallel Execution] Starting block execution",
		"block", blockNum,
		"txCount", txCount,
		"workers", a.workerCount)

	// Step 1: Extract header and transactions from tasks
	header, txs, senders, err := a.reconstructor.ExtractHeaderAndTransactions(tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to extract header and transactions: %w", err)
	}

	a.logger.Debug("[Parallel Execution] Extracted transactions",
		"block", blockNum,
		"txCount", len(txs),
		"senders", len(senders))

	// Step 2: Create StateDB adapter
	stateDB := NewStateDBAdapter(domains, tx, blockNum, txNum)

	// Step 3: Execute block in parallel
	result, err := a.executor.ProcessBlock(header, txs, stateDB)
	if err != nil {
		a.logger.Warn("[Parallel Execution] Execution failed",
			"block", blockNum,
			"error", err)
		return nil, fmt.Errorf("parallel execution failed: %w", err)
	}

	// Step 4: Log results
	if result != nil && result.Metrics != nil {
		a.logger.Info("[Parallel Execution] Block completed",
			"block", blockNum,
			"txCount", txCount,
			"gasUsed", result.GasUsed,
			"conflicts", result.Metrics.ConflictCount,
			"reexecutions", result.Metrics.ReexecutionCount,
			"duration", result.Metrics.TotalDuration,
			"speedup", result.Metrics.SpeedupFactor)
	} else {
		a.logger.Info("[Parallel Execution] Block completed",
			"block", blockNum,
			"txCount", txCount)
	}

	return result, nil
}

// GetMetrics returns execution metrics
func (a *ParallelExecutionAdapter) GetMetrics() *parallel.ExecutionMetrics {
	if !a.IsEnabled() {
		return nil
	}
	return a.executor.GetMetrics()
}
