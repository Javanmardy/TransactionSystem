package transaction_test

import (
	"TransactionSystem/internal/auth"
	"TransactionSystem/internal/transaction"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockTxService struct{}

func (m *mockTxService) GetTransactionByID(id int) *transaction.Transaction  { return nil }
func (m *mockTxService) AddTransaction(tx *transaction.Transaction) error    { return nil }
func (m *mockTxService) AllTransactions() ([]transaction.Transaction, error) { return nil, nil }
func (m *mockTxService) ListUserTransactions(userID int) []transaction.Transaction {
	return []transaction.Transaction{
		{ID: 1, UserID: userID, Amount: 123, Status: "success"},
		{ID: 2, UserID: userID, Amount: 456, Status: "failed"},
	}
}

func TestListUserTransactionsHandler(t *testing.T) {
	h := transaction.NewHandler(&mockTxService{})

	ctx := context.WithValue(context.Background(), auth.UserIDKey, 42)
	req := httptest.NewRequest("GET", "/transactions", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	h.ListUserTransactions(w, req)

	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	var got []transaction.Transaction
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("decode err: %v", err)
	}
	if len(got) != 2 || got[0].UserID != 42 {
		t.Errorf("unexpected tx list: %+v", got)
	}
}
