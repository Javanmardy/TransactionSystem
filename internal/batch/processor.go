package batch

import (
	"TransactionSystem/internal/transaction"
)

type BatchProcessor struct {
	strategies []ValidationStrategy
	service    transaction.Service
}

func NewBatchProcessor(service transaction.Service, strategies []ValidationStrategy) *BatchProcessor {
	return &BatchProcessor{
		strategies: strategies,
		service:    service,
	}
}

type BatchResult struct {
	ID     int    `json:"id,omitempty"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func (bp *BatchProcessor) Process(txs []transaction.Transaction, userID int) []BatchResult {
	var results []BatchResult
	for i := range txs {
		tx := &txs[i]
		var err error
		for _, strategy := range bp.strategies {
			if err = strategy.Validate(tx); err != nil {
				break
			}
		}
		if err != nil {
			results = append(results, BatchResult{
				Status: "failed",
				Error:  err.Error(),
			})
			continue
		}
		tx.UserID = userID
		_ = bp.service.AddTransaction(tx)
		results = append(results, BatchResult{
			ID:     tx.ID,
			Status: "success",
		})
	}
	return results
}
