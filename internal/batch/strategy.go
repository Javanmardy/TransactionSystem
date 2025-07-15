package batch

import (
	"TransactionSystem/internal/transaction"
	"errors"
)

type ValidationStrategy interface {
	Validate(tx *transaction.Transaction) error
}

type PositiveAmountStrategy struct{}

func (s *PositiveAmountStrategy) Validate(tx *transaction.Transaction) error {
	if tx.Amount <= 0 {
		return errors.New("amount must be positive")
	}
	return nil
}

type ValidStatusStrategy struct{}

func (s *ValidStatusStrategy) Validate(tx *transaction.Transaction) error {
	if tx.Status != "success" && tx.Status != "failed" {
		return errors.New("status must be 'success' or 'failed'")
	}
	return nil
}
