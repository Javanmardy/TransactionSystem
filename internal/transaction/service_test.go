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
