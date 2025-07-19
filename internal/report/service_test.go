package report

import (
	"TransactionSystem/internal/transaction"
	"testing"
)

func setupTestDB(t *testing.T) (transaction.Service, *transaction.DBRepo) {
	db, err := transaction.InitDB("root", "n61224n61224", "localhost:3306", "transaction_db")
	if err != nil {
		t.Fatalf("failed to connect to DB: %v", err)
	}
	repo := transaction.NewDBRepo(db)
	txService := transaction.NewService(repo)
	db.Exec("DELETE FROM transactions")
	return txService, repo
}

func TestUserReport(t *testing.T) {
	txService, repo := setupTestDB(t)
	reportService := NewService(txService)

	tx1 := &transaction.Transaction{UserID: 1, Amount: 1000, Status: "success"}
	tx2 := &transaction.Transaction{UserID: 1, Amount: 200, Status: "failed"}
	_ = txService.AddTransaction(tx1)
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

func TestAllReport(t *testing.T) {
	txService, repo := setupTestDB(t)
	reportService := NewService(txService)

	_ = txService.AddTransaction(&transaction.Transaction{UserID: 2, Amount: 400, Status: "success"})
	_ = txService.AddTransaction(&transaction.Transaction{UserID: 3, Amount: 800, Status: "failed"})

	report := reportService.AllReport()
	if report.TotalCount < 2 {
		t.Errorf("Expected at least 2 transactions, got %d", report.TotalCount)
	}
	if report.SuccessCount != 1 {
		t.Errorf("Expected 1 success, got %d", report.SuccessCount)
	}
	if report.FailedCount != 1 {
		t.Errorf("Expected 1 failed, got %d", report.FailedCount)
	}
	if report.TotalAmount < 1200 {
		t.Errorf("Expected total amount at least 1200, got %f", report.TotalAmount)
	}

	all, _ := txService.AllTransactions()
	for _, tx := range all {
		repo.DeleteTransaction(tx.ID)
	}
}
