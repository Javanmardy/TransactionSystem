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

func (bp *BatchProcessor) Process(txs []transaction.Transaction, actorID int) []BatchResult {
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
			results = append(results, BatchResult{Status: "failed", Error: err.Error()})
			continue
		}

		if tx.ToUserID == 0 && tx.UserID != 0 {
			tx.ToUserID = tx.UserID
		}
		if tx.UserID == 0 && tx.ToUserID != 0 {
			tx.UserID = tx.ToUserID
		}

		if actorID > 0 && tx.FromUserID == 0 {
			tx.FromUserID = actorID
		}

		if tx.UserID == 0 {
			results = append(results, BatchResult{Status: "failed", Error: "missing receiver user_id"})
			continue
		}

		if err := bp.service.AddTransaction(tx); err != nil {
			results = append(results, BatchResult{Status: "failed", Error: err.Error()})
			continue
		}
		results = append(results, BatchResult{ID: tx.ID, Status: "success"})
	}
	return results
}
