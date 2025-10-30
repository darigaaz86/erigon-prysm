package stagedsync

import (
	"fmt"

	"github.com/erigontech/erigon/common"
	"github.com/erigontech/erigon/execution/core/parallel"
	"github.com/erigontech/erigon/execution/exec"
	execTypes "github.com/erigontech/erigon/execution/types"
)

// BlockReconstructor extracts header and transactions from stagedsync tasks
type BlockReconstructor struct{}

// NewBlockReconstructor creates a new block reconstructor
func NewBlockReconstructor() *BlockReconstructor {
	return &BlockReconstructor{}
}

// ExtractHeaderAndTransactions extracts header and transactions from exec.Task array
// Returns header, transactions, and senders for parallel execution
func (br *BlockReconstructor) ExtractHeaderAndTransactions(tasks []exec.Task) (
	*execTypes.Header,
	[]execTypes.Transaction,
	[]common.Address,
	error,
) {
	if len(tasks) == 0 {
		return nil, nil, nil, fmt.Errorf("no tasks provided")
	}

	// Get the first task to extract block information
	firstTask := tasks[0].(*exec.TxTask)
	header := firstTask.Header

	// Extract transactions and senders from tasks
	var txs []execTypes.Transaction
	var senders []common.Address

	for _, task := range tasks {
		txTask := task.(*exec.TxTask)

		// Skip begin block (-1) and end block tasks
		if txTask.TxIndex < 0 || txTask.IsBlockEnd() {
			continue
		}

		tx := txTask.Tx()
		if tx == nil {
			continue
		}

		// Add transaction
		txs = append(txs, tx)

		// Extract sender - try to get from transaction
		sender, hasSender := tx.GetSender()
		if hasSender {
			senders = append(senders, sender)
		} else {
			// If sender not available, use zero address
			senders = append(senders, common.Address{})
		}
	}

	return header, txs, senders, nil
}

// WrapTransactions wraps execution/types.Transaction with senders for parallel execution
func (br *BlockReconstructor) WrapTransactions(
	txs []execTypes.Transaction,
	senders []common.Address,
) []parallel.Transaction {
	if len(txs) != len(senders) {
		// Mismatch - use zero addresses for missing senders
		if len(senders) < len(txs) {
			for i := len(senders); i < len(txs); i++ {
				senders = append(senders, common.Address{})
			}
		}
	}

	wrapped := make([]parallel.Transaction, len(txs))
	for i := range txs {
		wrapped[i] = parallel.NewTransactionWrapper(txs[i], senders[i], i)
	}

	return wrapped
}
