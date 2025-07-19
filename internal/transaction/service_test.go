package transaction_test

import (
	"TransactionSystem/internal/transaction"
	"testing"
)

func setupDB(t *testing.T) *transaction.DBRepo {
	db, err := transaction.InitDB("root", "n61224n61224", "localhost:3306", "transaction_db")
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}
	db.Exec("DELETE FROM transactions")
	return transaction.NewDBRepo(db)
}

func TestAddAndGetTransaction(t *testing.T) {
	repo := setupDB(t)
	svc := transaction.NewService(repo)

	tx := &transaction.Transaction{
		UserID: 1,
		Amount: 3500,
		Status: "success",
	}

	err := svc.AddTransaction(tx)
	if err != nil {
		t.Fatalf("Failed to add transaction: %v", err)
	}

	got := svc.GetTransactionByID(tx.ID)
	if got == nil || got.Amount != tx.Amount {
		t.Errorf("Expected amount %v, got %+v", tx.Amount, got)
	}

	repo.DeleteTransaction(tx.ID)
}

func TestListUserTransactions(t *testing.T) {
	repo := setupDB(t)
	svc := transaction.NewService(repo)

	_ = svc.AddTransaction(&transaction.Transaction{UserID: 2, Amount: 10, Status: "success"})
	_ = svc.AddTransaction(&transaction.Transaction{UserID: 2, Amount: 20, Status: "failed"})

	txList := svc.ListUserTransactions(2)
	if len(txList) < 2 {
		t.Errorf("expected at least 2 tx for user 2, got %d", len(txList))
	}
	for _, tx := range txList {
		repo.DeleteTransaction(tx.ID)
	}
}

func TestAllTransactions(t *testing.T) {
	repo := setupDB(t)
	svc := transaction.NewService(repo)

	_ = svc.AddTransaction(&transaction.Transaction{UserID: 5, Amount: 50, Status: "success"})
	_ = svc.AddTransaction(&transaction.Transaction{UserID: 6, Amount: 60, Status: "failed"})

	all, err := svc.AllTransactions()
	if err != nil {
		t.Fatalf("error on AllTransactions: %v", err)
	}
	if len(all) < 2 {
		t.Errorf("expected at least 2 tx, got %d", len(all))
	}

	for _, tx := range all {
		repo.DeleteTransaction(tx.ID)
	}
}
