package batch

import (
	"TransactionSystem/internal/auth"
	"TransactionSystem/internal/transaction"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("mysql", "root:n61224n61224@tcp(localhost:3306)/transaction_db")
	if err != nil {
		t.Fatalf("failed to connect db: %v", err)
	}

	_, err = db.Exec(`DELETE FROM transactions`)
	if err != nil {
		t.Fatalf("failed to clear transactions: %v", err)
	}
	return db
}

func TestProcessBatch_AdminOK_DB(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := transaction.NewDBRepo(db)
	svc := transaction.NewService(repo)
	h := NewHandler(svc)

	bodyObj := BatchRequest{
		Transactions: []transaction.Transaction{
			{UserID: 2, Amount: 5_000, Status: "success"},
			{UserID: 2, Amount: 300, Status: "failed"},
		},
	}
	body, _ := json.Marshal(bodyObj)

	req := httptest.NewRequest(http.MethodPost, "/batch", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), auth.RoleKey, "admin")
	ctx = context.WithValue(ctx, auth.UserIDKey, 1)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.ProcessBatch(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var res []BatchResult
	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 results, got %d", len(res))
	}
}

func TestProcessBatch_ForbiddenForNonAdmin_DB(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := transaction.NewDBRepo(db)
	svc := transaction.NewService(repo)
	h := NewHandler(svc)

	bodyObj := BatchRequest{
		Transactions: []transaction.Transaction{
			{Amount: 100, Status: "success"},
		},
	}
	body, _ := json.Marshal(bodyObj)
	req := httptest.NewRequest(http.MethodPost, "/batch", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), auth.RoleKey, "user")
	ctx = context.WithValue(ctx, auth.UserIDKey, 2)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.ProcessBatch(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestProcessBatch_InvalidData_DB(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := transaction.NewDBRepo(db)
	svc := transaction.NewService(repo)
	h := NewHandler(svc)

	bodyObj := BatchRequest{
		Transactions: []transaction.Transaction{
			{UserID: 3, Amount: -10, Status: "success"},
			{UserID: 3, Amount: 200, Status: "unknown"},
		},
	}
	body, _ := json.Marshal(bodyObj)

	req := httptest.NewRequest(http.MethodPost, "/batch", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), auth.RoleKey, "admin")
	ctx = context.WithValue(ctx, auth.UserIDKey, 1)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	h.ProcessBatch(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var res []BatchResult
	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 results, got %d", len(res))
	}
	for _, r := range res {
		if r.Status != "failed" || r.Error == "" {
			t.Errorf("expected failed status and error, got %+v", r)
		}
	}
}

func TestProcessBatch_BadRequest_DB(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := transaction.NewDBRepo(db)
	svc := transaction.NewService(repo)
	h := NewHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/batch", bytes.NewReader([]byte("not-json")))
	ctx := context.WithValue(req.Context(), auth.RoleKey, "admin")
	ctx = context.WithValue(ctx, auth.UserIDKey, 1)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	h.ProcessBatch(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
