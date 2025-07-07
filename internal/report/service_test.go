package report

import (
	"TransactionSystem/internal/transaction"
	"testing"
)

func TestUserReport(t *testing.T) {
	mockRepo := transaction.NewMockRepo()
	txService := transaction.NewService(mockRepo)
	reportService := NewService(txService)

	report := reportService.UserReport(1)
	if report.TotalCount != 2 {
		t.Errorf("Expected 2 transactions, got %d", report.TotalCount)
	}
	if report.SuccessCount != 1 {
		t.Errorf("Expected 1 successful transaction, got %d", report.SuccessCount)
	}
	if report.FailedCount != 1 {
		t.Errorf("Expected 1 failed transaction, got %d", report.FailedCount)
	}
	if report.SuccessRate != 0.5 {
		t.Errorf("Expected success rate 0.5, got %f", report.SuccessRate)
	}
}
