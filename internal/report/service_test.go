package report

import (
	"TransactionSystem/internal/transaction"
	"testing"
)

func TestUserReport(t *testing.T) {
	db, err := transaction.InitDB("root", "n61224n61224", "localhost:3306", "transaction_db")
	if err != nil {
		t.Fatalf("failed to connect to DB: %v", err)
	}
	repo := transaction.NewDBRepo(db)
	txService := transaction.NewService(repo)
	reportService := NewService(txService)

	tx1 := &transaction.Transaction{
		UserID: 1,
		Amount: 1000,
		Status: "success",
	}
	_ = txService.AddTransaction(tx1)

	tx2 := &transaction.Transaction{
		UserID: 1,
		Amount: 200,
		Status: "failed",
	}
	_ = txService.AddTransaction(tx2)

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

	repo.DeleteTransaction(tx1.ID)
	repo.DeleteTransaction(tx2.ID)
}
