// Package parallel implements Block-STM parallel transaction execution for Erigon.
//
// This package provides a multi-version concurrency control (MVCC) based approach
// to execute transactions in parallel across multiple CPU cores while maintaining
// deterministic results.
//
// Architecture:
//   - MVCC State Manager: Maintains multiple versions of state variables
//   - Worker Pool: Executes transactions concurrently
//   - Conflict Validator: Detects read-write conflicts
//   - Re-execution Scheduler: Handles invalid transactions
//   - Commit Manager: Applies validated writes to final state
//
// Usage:
//
//	executor := parallel.NewParallelExecutor(config)
//	result, err := executor.ExecuteBlock(block, state)
package parallel
