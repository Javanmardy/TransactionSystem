package transaction_test

import (
	"TransactionSystem/internal/auth"
	"TransactionSystem/internal/transaction"
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockTxServiceOK struct{}

func (m *mockTxServiceOK) GetTransactionByID(id int) *transaction.Transaction  { return nil }
func (m *mockTxServiceOK) AddTransaction(tx *transaction.Transaction) error    { return nil }
func (m *mockTxServiceOK) AllTransactions() ([]transaction.Transaction, error) { return nil, nil }
func (m *mockTxServiceOK) TransferFunds(fromUserID, toUserID int, amount float64, status string) error {
	return nil
}
func (m *mockTxServiceOK) ListUserTransactions(userID int) []transaction.Transaction {
	return []transaction.Transaction{{ID: 1, UserID: userID, Amount: 100, Status: "success"}}
}

type mockTxServiceErr struct{ mockTxServiceOK }

func (m *mockTxServiceErr) TransferFunds(fromUserID, toUserID int, amount float64, status string) error {
	return errors.New("transfer failed")
}

func TestListUserTransactions_Unauthorized(t *testing.T) {
	h := transaction.NewHandler(&mockTxServiceOK{})
	req := httptest.NewRequest("GET", "/transactions", nil)
	w := httptest.NewRecorder()

	h.ListUserTransactions(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAddTransactionHandler_Success(t *testing.T) {
	h := transaction.NewHandler(&mockTxServiceOK{})
	ctx := context.WithValue(context.Background(), auth.UserIDKey, 1)
	body := bytes.NewBufferString(`{"to_user_id":2,"amount":50}`)
	req := httptest.NewRequest("POST", "/transactions", body).WithContext(ctx)
	w := httptest.NewRecorder()

	h.AddTransactionHandler(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestAddTransactionHandler_InvalidRecipient(t *testing.T) {
	h := transaction.NewHandler(&mockTxServiceOK{})
	ctx := context.WithValue(context.Background(), auth.UserIDKey, 1)
	body := bytes.NewBufferString(`{"to_user_id":1,"amount":50}`)
	req := httptest.NewRequest("POST", "/transactions", body).WithContext(ctx)
	w := httptest.NewRecorder()

	h.AddTransactionHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAddTransactionHandler_InvalidAmount(t *testing.T) {
	h := transaction.NewHandler(&mockTxServiceOK{})
	ctx := context.WithValue(context.Background(), auth.UserIDKey, 1)
	body := bytes.NewBufferString(`{"to_user_id":2,"amount":0}`)
	req := httptest.NewRequest("POST", "/transactions", body).WithContext(ctx)
	w := httptest.NewRecorder()

	h.AddTransactionHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAddTransactionHandler_TransferError(t *testing.T) {
	h := transaction.NewHandler(&mockTxServiceErr{})
	ctx := context.WithValue(context.Background(), auth.UserIDKey, 1)
	body := bytes.NewBufferString(`{"to_user_id":2,"amount":50}`)
	req := httptest.NewRequest("POST", "/transactions", body).WithContext(ctx)
	w := httptest.NewRecorder()

	h.AddTransactionHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAddTransactionHandler_InvalidJSON(t *testing.T) {
	h := transaction.NewHandler(&mockTxServiceOK{})
	ctx := context.WithValue(context.Background(), auth.UserIDKey, 1)
	body := bytes.NewBufferString(`{"to_user_id":}`)
	req := httptest.NewRequest("POST", "/transactions", body).WithContext(ctx)
	w := httptest.NewRecorder()

	h.AddTransactionHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
