package parallel

import (
	"fmt"

	"github.com/erigontech/erigon/execution/types"
)

// SequentialExecutor executes transactions sequentially (fallback)
type SequentialExecutor struct{}

// NewSequentialExecutor creates a new sequential executor
func NewSequentialExecutor() *SequentialExecutor {
	return &SequentialExecutor{}
}

// ExecuteBlock executes transactions sequentially
func (se *SequentialExecutor) ExecuteBlock(
	header *types.Header,
	txs []types.Transaction,
	state StateDB,
) (*ExecutionResult, error) {
	// TODO: Implement sequential execution
	// For now, return error to indicate not implemented
	return nil, fmt.Errorf("sequential execution not yet implemented")
}
