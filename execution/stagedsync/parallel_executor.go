package stagedsync

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/erigontech/erigon/common"
	"github.com/erigontech/erigon/common/log/v3"
	"github.com/erigontech/erigon/execution/exec"
	"github.com/erigontech/erigon/execution/exec3"
)

// ParallelTaskExecutor executes stagedsync tasks in parallel
type ParallelTaskExecutor struct {
	logger      log.Logger
	workerCount int
	enabled     bool
}

// NewParallelTaskExecutor creates a new parallel task executor
func NewParallelTaskExecutor(logger log.Logger, workerCount int) *ParallelTaskExecutor {
	return &ParallelTaskExecutor{
		logger:      logger,
		workerCount: workerCount,
		enabled:     true,
	}
}

// TaskResult holds the result of executing a task
type TaskResult struct {
	TaskIndex int
	TxIndex   int
	Result    *exec.TxResult
	Error     error
}

// ExecuteTasks executes tasks in parallel
func (pte *ParallelTaskExecutor) ExecuteTasks(
	ctx context.Context,
	tasks []exec.Task,
	worker *exec3.Worker,
) ([]TaskResult, error) {
	if !pte.enabled {
		return nil, fmt.Errorf("parallel execution is disabled")
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("no tasks to execute")
	}

	startTime := time.Now()

	// Count actual transaction tasks
	txCount := 0
	for _, task := range tasks {
		txTask := task.(*exec.TxTask)
		if txTask.TxIndex >= 0 && !txTask.IsBlockEnd() {
			txCount++
		}
	}

	pte.logger.Info("[Parallel Execution] Starting parallel execution",
		"txCount", txCount,
		"workers", pte.workerCount)

	// Create channels for work distribution
	taskChan := make(chan exec.Task, len(tasks))
	resultChan := make(chan TaskResult, len(tasks))

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < pte.workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			pte.executeWorker(ctx, workerID, taskChan, resultChan, worker)
		}(i)
	}

	// Send tasks to workers
	go func() {
		for i, task := range tasks {
			select {
			case taskChan <- task:
				_ = i // task index
			case <-ctx.Done():
				return
			}
		}
		close(taskChan)
	}()

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	results := make([]TaskResult, 0, len(tasks))
	for result := range resultChan {
		results = append(results, result)
	}

	duration := time.Since(startTime)
	pte.logger.Info("[Parallel Execution] Completed parallel execution",
		"txCount", txCount,
		"duration", duration,
		"results", len(results))

	return results, nil
}

// executeWorker is a worker goroutine that processes tasks
func (pte *ParallelTaskExecutor) executeWorker(
	ctx context.Context,
	workerID int,
	taskChan <-chan exec.Task,
	resultChan chan<- TaskResult,
	worker *exec3.Worker,
) {
	for {
		select {
		case task, ok := <-taskChan:
			if !ok {
				return // Channel closed, worker done
			}

			txTask := task.(*exec.TxTask)

			// Skip begin/end block tasks
			if txTask.TxIndex < 0 || txTask.IsBlockEnd() {
				continue
			}

			// Execute the task
			result := worker.RunTxTask(txTask)

			// Send result
			resultChan <- TaskResult{
				TxIndex: txTask.TxIndex,
				Result:  result,
				Error:   result.Err,
			}

		case <-ctx.Done():
			return
		}
	}
}

// ShouldUseParallel determines if parallel execution should be used
func (pte *ParallelTaskExecutor) ShouldUseParallel(txCount int) bool {
	return pte.enabled && txCount >= 10
}

// ConflictDetector detects conflicts between transaction executions
type ConflictDetector struct {
	logger log.Logger
}

// NewConflictDetector creates a new conflict detector
func NewConflictDetector(logger log.Logger) *ConflictDetector {
	return &ConflictDetector{
		logger: logger,
	}
}

// ReadWriteSet tracks reads and writes for a transaction
type ReadWriteSet struct {
	Reads  map[common.Address]map[common.Hash]bool // address -> storage keys
	Writes map[common.Address]map[common.Hash]bool // address -> storage keys
}

// DetectConflicts detects conflicts between transaction results
func (cd *ConflictDetector) DetectConflicts(results []TaskResult) []int {
	// For now, assume no conflicts
	// A full implementation would track read/write sets and detect overlaps
	return nil
}

// ParallelExecutionMetrics tracks metrics for parallel execution
type ParallelExecutionMetrics struct {
	BlocksProcessed   int64
	TxsProcessed      int64
	ConflictsDetected int64
	Reexecutions      int64
	TotalDuration     time.Duration
	AverageSpeedup    float64
}

// NewParallelExecutionMetrics creates new metrics
func NewParallelExecutionMetrics() *ParallelExecutionMetrics {
	return &ParallelExecutionMetrics{}
}

// RecordExecution records a parallel execution
func (m *ParallelExecutionMetrics) RecordExecution(txCount int, duration time.Duration, conflicts int) {
	m.BlocksProcessed++
	m.TxsProcessed += int64(txCount)
	m.ConflictsDetected += int64(conflicts)
	m.TotalDuration += duration
}
