package batch

import (
	"TransactionSystem/internal/transaction"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProcessBatch(t *testing.T) {
	db, err := transaction.InitDB("root", "n61224n61224", "localhost:3306", "transaction_db")
	if err != nil {
		t.Fatalf("failed to connect to DB: %v", err)
	}
	repo := transaction.NewDBRepo(db)
	svc := transaction.NewService(repo)
	handler := NewHandler(svc)

	reqBody := BatchRequest{
		Transactions: []transaction.Transaction{
			{Amount: 500, Status: "success"},
			{Amount: -100, Status: "failed"},
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/batch", bytes.NewReader(body))

	ctx := context.WithValue(req.Context(), "userID", 1)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ProcessBatch(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d", rec.Code)
	}

	var results []BatchResult
	if err := json.Unmarshal(rec.Body.Bytes(), &results); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Status != "success" {
		t.Errorf("expected first status to be success, got %s", results[0].Status)
	}
	if results[1].Status != "failed" || results[1].Error != "amount must be positive" {
		t.Errorf("expected error for negative amount, got %+v", results[1])
	}
}
