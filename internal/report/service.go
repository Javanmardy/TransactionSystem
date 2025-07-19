package report

import (
	"TransactionSystem/internal/transaction"
)

type Report struct {
	TotalCount    int
	SuccessCount  int
	FailedCount   int
	TotalAmount   float64
	SuccessAmount float64
	FailedAmount  float64
	SuccessRate   float64
}

type Service interface {
	UserReport(userID int) Report
	AllTransactions() ([]transaction.Transaction, error)
}

type service struct {
	txService transaction.Service
}

func NewService(txSvc transaction.Service) Service {
	return &service{txService: txSvc}
}

func (s *service) UserReport(userID int) Report {
	txs := s.txService.ListUserTransactions(userID)
	var rep Report
	rep.TotalCount = len(txs)
	for _, t := range txs {
		rep.TotalAmount += t.Amount
		if t.Status == "success" {
			rep.SuccessCount++
			rep.SuccessAmount += t.Amount
		} else {
			rep.FailedCount++
			rep.FailedAmount += t.Amount
		}
	}
	if rep.TotalCount > 0 {
		rep.SuccessRate = float64(rep.SuccessCount) / float64(rep.TotalCount)
	}
	return rep
}
func (s *service) AllTransactions() ([]transaction.Transaction, error) {
	return s.txService.AllTransactions()
}
