package parallel

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/erigontech/erigon/common"
	log "github.com/erigontech/erigon/common/log/v3"
	"github.com/erigontech/erigon/execution/types"
	"github.com/erigontech/erigon/execution/core/parallel/commit"
	"github.com/erigontech/erigon/execution/core/parallel/mvcc"
	"github.com/erigontech/erigon/execution/core/parallel/scheduler"
	"github.com/erigontech/erigon/execution/core/parallel/validator"
	"github.com/erigontech/erigon/execution/core/parallel/worker"
)

// parallelExecutor implements the ParallelExecutor interface using production types.
type parallelExecutor struct {
	config            *Config
	workerPool        worker.WorkerPool
	mvccManager       mvcc.MVCCStateManager
	conflictValidator validator.ConflictValidator
	reexecScheduler   scheduler.ReexecutionScheduler
	commitManager     commit.CommitManager
	metrics           *ExecutionMetrics
}

// NewParallelExecutor creates a new parallel executor using production types.
func NewParallelExecutor(config *Config) ParallelExecutor {
	if config == nil {
		config = DefaultConfig()
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		// Use default config if validation fails
		config = DefaultConfig()
	}

	// Create MVCC manager
	mvccManager := mvcc.NewManager(nil)

	// Create components
	workerPool := worker.NewWorkerPool(config.WorkerCount, 100)
	conflictValidator := validator.NewConflictValidator()
	reexecScheduler := scheduler.NewReexecutionScheduler(
		&scheduler.ReexecutionConfig{
			MaxRetries: config.MaxReexecutions,
			Strategy:   scheduler.SequentialReexecution,
			BatchSize:  10,
		},
		mvccManager,
	)
	commitManager := commit.NewCommitManager()

	return &parallelExecutor{
		config:            config,
		workerPool:        workerPool,
		mvccManager:       mvccManager,
		conflictValidator: conflictValidator,
		reexecScheduler:   reexecScheduler,
		commitManager:     commitManager,
		metrics:           NewExecutionMetrics(),
	}
}

// ExecuteBlock executes all transactions in a block in parallel using production types.
func (pe *parallelExecutor) ExecuteBlock(
	header *types.Header,
	txs []types.Transaction,
	state StateDB,
) (*ExecutionResult, error) {
	if !pe.config.Enabled {
		// Parallel execution is disabled, fall back to sequential
		log.Warn("[Parallel Execution] Disabled, falling back to sequential")
		return nil, fmt.Errorf("parallel execution is disabled")
	}

	// log.Info("[Parallel Execution] ExecuteBlock called", "block", header.Number, "txCount", len(txs))

	startTime := time.Now()

	// Reset state for new block
	pe.mvccManager.Reset()
	pe.conflictValidator.Reset()
	pe.reexecScheduler.Reset()
	pe.commitManager.Reset()

	// log.Debug("[Parallel Execution] State reset complete")

	// Wrap transactions
	wrappedTxs := make([]Transaction, len(txs))
	for i, tx := range txs {
		// TODO: Extract sender from transaction
		sender := common.Address{} // Placeholder
		wrappedTxs[i] = NewTransactionWrapper(tx, sender, i)
	}

	// log.Debug("[Parallel Execution] Transactions wrapped", "count", len(wrappedTxs))

	// Phase 1: Parallel Execution
	// log.Info("[Parallel Execution] Phase 1: Starting parallel execution")
	results, err := pe.executeParallel(header, wrappedTxs, state)
	if err != nil {
		log.Error("[Parallel Execution] Phase 1 failed", "error", err)
		return nil, fmt.Errorf("parallel execution failed: %w", err)
	}

	executionTime := time.Since(startTime)
	// log.Info("[Parallel Execution] Phase 1 complete", "duration", executionTime, "results", len(results))

	// Phase 2: Validation
	// log.Info("[Parallel Execution] Phase 2: Starting validation")
	validationStart := time.Now()
	// Convert to validator.TxResult (interface conversion)
	validatorResults := make([]interface{}, len(results))
	for i, r := range results {
		validatorResults[i] = r
	}
	invalidTxs := pe.conflictValidator.Validate(validatorResults)
	validationTime := time.Since(validationStart)
	// log.Info("[Parallel Execution] Phase 2 complete", "duration", validationTime, "conflicts", len(invalidTxs))

	// Phase 3: Re-execution (if needed)
	// log.Info("[Parallel Execution] Phase 3: Re-execution check", "invalidTxs", len(invalidTxs))
	reexecStart := time.Now()
	reexecCount := 0
	if len(invalidTxs) > 0 {
		// log.Info("[Parallel Execution] Re-executing conflicting transactions", "count", len(invalidTxs))
		// Convert to scheduler.TxResult (interface conversion)
		schedulerResults := make([]interface{}, len(results))
		for i, r := range results {
			schedulerResults[i] = r
		}
		if err := pe.reexecScheduler.Reexecute(invalidTxs, schedulerResults); err != nil {
			log.Error("[Parallel Execution] Phase 3 failed", "error", err)
			return nil, fmt.Errorf("re-execution failed: %w", err)
		}
		reexecCount = len(invalidTxs)
		// log.Info("[Parallel Execution] Phase 3 complete", "reexecuted", reexecCount)
	} else {
		// log.Info("[Parallel Execution] Phase 3 skipped - no conflicts")
	}
	reexecTime := time.Since(reexecStart)

	// Phase 4: Commit
	// log.Info("[Parallel Execution] Phase 4: Starting commit")
	commitStart := time.Now()
	// Commit expects []*commit.TxResult, but we have []*TxResult
	// Since they're compatible (same structure), we can use interface{} cast
	// log.Debug("[Parallel Execution] Calling commit manager", "results", len(results))
	commitResult, err := pe.commitManager.Commit(results, state)
	if err != nil {
		log.Error("[Parallel Execution] Phase 4 failed", "error", err)
		return nil, fmt.Errorf("commit failed: %w", err)
	}
	commitTime := time.Since(commitStart)
	// log.Info("[Parallel Execution] Phase 4 complete", "duration", commitTime)
	
	// Cast result back to ExecutionResult
	result, ok := commitResult.(*ExecutionResult)
	if !ok {
		// Try to get it as interface and convert
		if execResult, ok := commitResult.(ExecutionResult); ok {
			result = &execResult
		} else {
			log.Error("[Parallel Execution] Invalid commit result type")
			return nil, fmt.Errorf("invalid commit result type")
		}
	}

	// Update metrics
	totalTime := time.Since(startTime)
	pe.updateMetrics(len(txs), reexecCount, executionTime, validationTime, reexecTime, commitTime, totalTime)

	// Add metrics to result
	result.Metrics = pe.metrics

	// log.Info("[Parallel Execution] Block execution complete", 
	// 	"block", header.Number,
	// 	"txs", len(txs),
	// 	"conflicts", reexecCount,
	// 	"totalTime", totalTime,
	// 	"speedup", fmt.Sprintf("%.2fx", pe.metrics.SpeedupFactor))

	return result, nil
}

// executeParallel executes transactions in parallel
func (pe *parallelExecutor) executeParallel(
	header *types.Header,
	txs []Transaction,
	state StateDB,
) ([]*TxResult, error) {
	// log.Info("[Parallel Execution] Starting parallel execution", "txCount", len(txs), "workers", pe.config.WorkerCount, "block", header.Number)

	// Start worker pool
	ctx := context.Background()
	pe.workerPool.Start(ctx)
	defer pe.workerPool.Stop()

	// log.Debug("[Parallel Execution] Worker pool started")

	// Create MVCC state wrappers for each transaction
	stateWrappers := mvcc.CreateStateWrappers(pe.mvccManager, state, len(txs))
	// log.Debug("[Parallel Execution] Created MVCC state wrappers", "count", len(stateWrappers))

	// Submit tasks to workers
	// log.Debug("[Parallel Execution] Submitting tasks to workers")
	for i, tx := range txs {
		task := &worker.Task{
			TxIndex:     i,
			Transaction: tx,
			State:       stateWrappers[i],
			Header:      header,
		}
		pe.workerPool.SubmitTask(task)
	}

	// Collect results
	// log.Debug("[Parallel Execution] Collecting results from workers")
	results := make([]*TxResult, 0, len(txs))
	resultChan := pe.workerPool.GetResults()

	successCount := 0
	errorCount := 0

	for i := 0; i < len(txs); i++ {
		workerResult := <-resultChan

		if workerResult.Error != nil {
			errorCount++
			// log.Debug("[Parallel Execution] Transaction failed", "txIndex", workerResult.TxIndex, "error", workerResult.Error)
		} else {
			successCount++
		}

		// Convert worker.Result to TxResult
		txResult := &TxResult{
			TxIndex:     workerResult.TxIndex,
			TxHash:      workerResult.TxHash,
			Receipt:     workerResult.Receipt,
			Logs:        workerResult.Logs,
			GasUsed:     workerResult.GasUsed,
			BlobGasUsed: workerResult.BlobGasUsed,
			ReadSet:     workerResult.ReadSet,
			WriteSet:    workerResult.WriteSet,
			ReexecCount: 0,
			Error:       workerResult.Error,
		}

		results = append(results, txResult)
	}

	// log.Info("[Parallel Execution] Collected all results", "total", len(results), "success", successCount, "errors", errorCount)

	// Sort results by transaction index
	sort.Slice(results, func(i, j int) bool {
		return results[i].TxIndex < results[j].TxIndex
	})

	return results, nil
}

// updateMetrics updates execution metrics
func (pe *parallelExecutor) updateMetrics(
	totalTxs, reexecCount int,
	execTime, valTime, reexecTime, commitTime, totalTime time.Duration,
) {
	pe.metrics.TotalTxs = totalTxs
	pe.metrics.ParallelTxs = totalTxs
	pe.metrics.ReexecutedTxs = reexecCount
	pe.metrics.ExecutionTime = execTime
	pe.metrics.ValidationTime = valTime
	pe.metrics.ReexecutionTime = reexecTime
	pe.metrics.CommitTime = commitTime
	pe.metrics.TotalDuration = totalTime

	if reexecCount > 0 {
		pe.metrics.ConflictRate = float64(reexecCount) / float64(totalTxs)
	}

	// Calculate speedup (placeholder - will be more accurate with real measurements)
	pe.metrics.SpeedupFactor = float64(pe.config.WorkerCount) * 0.7 // Rough estimate
}

// SetWorkerCount configures the number of worker threads
func (pe *parallelExecutor) SetWorkerCount(count int) {
	pe.config.WorkerCount = count
	// TODO: Recreate worker pool with new count
}

// GetMetrics returns execution metrics
func (pe *parallelExecutor) GetMetrics() *ExecutionMetrics {
	return pe.metrics
}


