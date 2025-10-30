package worker

import (
	"context"
	"fmt"
	"sync"

	"github.com/erigontech/erigon/common"
	"github.com/erigontech/erigon/execution/core/parallel/mvcc"
	"github.com/erigontech/erigon/execution/types"
	"github.com/erigontech/erigon/execution/types/accounts"
	"github.com/holiman/uint256"
)

// Transaction interface (duplicated to avoid import cycle)
type Transaction interface {
	Hash() common.Hash
	Type() byte
	GetSender() (common.Address, error)
	GetTo() *common.Address
	GetValue() *uint256.Int
	GetData() []byte
	GetGas() uint64
	GetPrice() *uint256.Int
	GetNonce() uint64
	GetChainID() *uint256.Int
	IsContractCreation() bool
	Protected() bool
	GetAccessList() types.AccessList
	GetBlobHashes() []common.Hash
	GetBlobGas() uint64
}

// StateDB interface (duplicated to avoid import cycle)
type StateDB interface {
	GetAccount(addr common.Address) (*accounts.Account, error)
	SetAccount(addr common.Address, account *accounts.Account) error
	GetState(addr common.Address, key common.Hash) (common.Hash, error)
	SetState(addr common.Address, key, value common.Hash) error
	GetCode(addr common.Address) ([]byte, error)
	SetCode(addr common.Address, code []byte) error
	GetBalance(addr common.Address) (*uint256.Int, error)
	SetBalance(addr common.Address, balance *uint256.Int) error
	GetNonce(addr common.Address) (uint64, error)
	SetNonce(addr common.Address, nonce uint64) error
}

// Worker executes transactions in parallel using MVCC state
type Worker struct {
	id         int
	taskChan   chan *Task
	resultChan chan *Result
	stopChan   chan struct{}
	wg         *sync.WaitGroup
}

// NewWorker creates a new worker
func NewWorker(id int, taskChan chan *Task, resultChan chan *Result, wg *sync.WaitGroup) *Worker {
	return &Worker{
		id:         id,
		taskChan:   taskChan,
		resultChan: resultChan,
		stopChan:   make(chan struct{}),
		wg:         wg,
	}
}

// Start begins the worker's execution loop
func (w *Worker) Start(ctx context.Context) {
	w.wg.Add(1)
	go w.run(ctx)
}

// Stop signals the worker to stop
func (w *Worker) Stop() {
	close(w.stopChan)
}

// run is the main worker loop
func (w *Worker) run(ctx context.Context) {
	defer w.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopChan:
			return
		case task, ok := <-w.taskChan:
			if !ok {
				return
			}
			result := w.executeTask(ctx, task)
			w.resultChan <- result
		}
	}
}

// executeTask executes a single transaction task
func (w *Worker) executeTask(ctx context.Context, task *Task) *Result {
	result := &Result{
		TxIndex: task.TxIndex,
		TxHash:  task.Transaction.Hash(),
	}

	// Execute the transaction with MVCC state
	receipt, logs, gasUsed, blobGasUsed, err := w.executeTransaction(ctx, task)
	if err != nil {
		result.Error = err
		return result
	}

	result.Receipt = receipt
	result.Logs = logs
	result.GasUsed = gasUsed
	result.BlobGasUsed = blobGasUsed

	// Get read and write sets from MVCC state
	if mvccState, ok := task.State.(mvcc.MVCCStateDB); ok {
		result.ReadSet = mvccState.GetReadSet()
		result.WriteSet = mvccState.GetWriteSet()
	}

	return result
}

// executeTransaction executes a transaction and returns the result
// This is a simplified implementation that focuses on state tracking
func (w *Worker) executeTransaction(
	ctx context.Context,
	task *Task,
) (*types.Receipt, []*types.Log, uint64, uint64, error) {
	// For now, we'll create a minimal implementation that:
	// 1. Tracks state accesses through the MVCC state wrapper
	// 2. Returns a basic receipt
	// 3. Calculates gas usage
	
	// In a full implementation, this would:
	// - Create an EVM instance
	// - Execute the transaction through the EVM
	// - Generate proper receipts and logs
	// - Handle all transaction types (legacy, EIP-1559, EIP-4844, etc.)
	
	tx := task.Transaction
	
	// Get sender
	sender, err := tx.GetSender()
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("failed to get sender: %w", err)
	}
	
	// Basic state operations to track reads/writes
	// This simulates what the EVM would do
	
	// 1. Check sender balance and nonce
	senderBalance, err := task.State.GetBalance(sender)
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("failed to get sender balance: %w", err)
	}
	
	senderNonce, err := task.State.GetNonce(sender)
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("failed to get sender nonce: %w", err)
	}
	
	// 2. Validate nonce
	if senderNonce != tx.GetNonce() {
		return nil, nil, 0, 0, fmt.Errorf("invalid nonce: got %d, expected %d", tx.GetNonce(), senderNonce)
	}
	
	// 3. Calculate gas cost
	gasPrice := tx.GetPrice()
	gasLimit := tx.GetGas()
	gasCost := uint256.NewInt(0).Mul(gasPrice, uint256.NewInt(gasLimit))
	
	// 4. Check if sender has enough balance
	totalCost := uint256.NewInt(0).Add(gasCost, tx.GetValue())
	if senderBalance.Cmp(totalCost) < 0 {
		return nil, nil, 0, 0, fmt.Errorf("insufficient balance: have %s, need %s", senderBalance.String(), totalCost.String())
	}
	
	// 5. Update sender state
	newBalance := uint256.NewInt(0).Sub(senderBalance, totalCost)
	if err := task.State.SetBalance(sender, newBalance); err != nil {
		return nil, nil, 0, 0, fmt.Errorf("failed to set sender balance: %w", err)
	}
	
	newNonce := senderNonce + 1
	if err := task.State.SetNonce(sender, newNonce); err != nil {
		return nil, nil, 0, 0, fmt.Errorf("failed to set sender nonce: %w", err)
	}
	
	// 6. Handle recipient (if not contract creation)
	recipient := tx.GetTo()
	if recipient != nil {
		// Transfer value to recipient
		recipientBalance, err := task.State.GetBalance(*recipient)
		if err != nil {
			return nil, nil, 0, 0, fmt.Errorf("failed to get recipient balance: %w", err)
		}
		
		newRecipientBalance := uint256.NewInt(0).Add(recipientBalance, tx.GetValue())
		if err := task.State.SetBalance(*recipient, newRecipientBalance); err != nil {
			return nil, nil, 0, 0, fmt.Errorf("failed to set recipient balance: %w", err)
		}
	}
	
	// 7. Create receipt (simplified)
	receipt := &types.Receipt{
		Type:              tx.Type(),
		Status:            types.ReceiptStatusSuccessful,
		CumulativeGasUsed: gasLimit, // Simplified: assume all gas used
		TxHash:            tx.Hash(),
		GasUsed:           gasLimit,
	}
	
	// 8. Create logs (empty for simple transfers)
	logs := make([]*types.Log, 0)
	
	// 9. Calculate blob gas (if applicable)
	blobGasUsed := tx.GetBlobGas()
	
	return receipt, logs, gasLimit, blobGasUsed, nil
}

// Task represents a transaction execution task
type Task struct {
	TxIndex     int
	Transaction Transaction
	State       StateDB
	Header      *types.Header
	ChainConfig interface{} // Placeholder for chain config
	VMConfig    interface{} // Placeholder for VM config
}

// Result represents the result of executing a transaction
type Result struct {
	TxIndex     int
	TxHash      common.Hash
	Receipt     *types.Receipt
	Logs        []*types.Log
	GasUsed     uint64
	BlobGasUsed uint64
	ReadSet     *mvcc.ReadSet
	WriteSet    *mvcc.WriteSet
	Error       error
}


