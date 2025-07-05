package report

import "TransactionSystem/internal/transaction"

type Report struct {
	UserID            int
	TotalTransactions int
	SuccessfulCount   int
	FailedCount       int
	TotalAmount       float64
}

type Service interface {
	GenerateUserReport(userID int) *Report
}

type reportService struct {
	transactionSvc transaction.Service
}

func NewService(transactionSvc transaction.Service) Service {
	return &reportService{
		transactionSvc: transactionSvc,
	}
}

func (s *reportService) GenerateUserReport(userID int) *Report {
	transactions := s.transactionSvc.ListUserTransactions(userID)

	var success, fail int
	var amount float64

	for _, t := range transactions {
		if t.Status == "success" {
			success++
		} else {
			fail++
		}
		amount += t.Amount
	}

	return &Report{
		UserID:            userID,
		TotalTransactions: len(transactions),
		SuccessfulCount:   success,
		FailedCount:       fail,
		TotalAmount:       amount,
	}
}
