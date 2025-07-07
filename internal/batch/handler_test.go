package batch

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"TransactionSystem/internal/transaction"
)

func TestProcessBatch(t *testing.T) {
	mockRepo := transaction.NewMockRepo()
	txService := transaction.NewService(mockRepo)
	handler := NewHandler(txService)

	reqBody := BatchRequest{
		UserID: 1,
		Transactions: []transaction.Transaction{
			{ID: 100, Amount: 1000, Status: "pending"},
			{ID: 101, Amount: -50, Status: "pending"},
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ProcessBatch(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status OK, got %d", rr.Code)
	}

	var res []BatchResult
	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatalf("Could not parse response: %v", err)
	}

	if len(res) != 2 {
		t.Errorf("Expected 2 batch results, got %d", len(res))
	}
	if res[0].Status != "success" {
		t.Errorf("Expected first transaction to succeed, got %s", res[0].Status)
	}
	if res[1].Status != "failed" || res[1].Error != "amount must be positive" {
		t.Errorf("Expected second transaction to fail with error 'amount must be positive', got status %s error %s", res[1].Status, res[1].Error)
	}
}
