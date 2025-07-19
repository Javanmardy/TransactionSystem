package report_test

import (
	"TransactionSystem/internal/auth"
	"TransactionSystem/internal/report"
	"TransactionSystem/internal/transaction"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockReportService struct{}

func (m *mockReportService) UserReport(userID int) report.Report {
	return report.Report{
		TotalCount: 2, SuccessCount: 1, FailedCount: 1,
		TotalAmount: 1000, SuccessAmount: 800, FailedAmount: 200, SuccessRate: 0.5,
	}
}
func (m *mockReportService) AllTransactions() ([]transaction.Transaction, error) {
	return []transaction.Transaction{
		{ID: 1, UserID: 1, Amount: 500, Status: "success"},
		{ID: 2, UserID: 2, Amount: 200, Status: "failed"},
	}, nil
}
func (m *mockReportService) AllReport() report.Report {
	return report.Report{
		TotalCount: 5, SuccessCount: 3, FailedCount: 2, TotalAmount: 700,
	}
}

func TestUserReportHandler(t *testing.T) {
	h := report.NewHandler(&mockReportService{})

	ctx := context.WithValue(context.Background(), auth.UserIDKey, 7)
	req := httptest.NewRequest("GET", "/report", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	h.UserReport(w, req)

	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	var rep report.Report
	if err := json.NewDecoder(res.Body).Decode(&rep); err != nil {
		t.Fatalf("decode err: %v", err)
	}
	if rep.TotalCount != 2 || rep.SuccessCount != 1 || rep.FailedCount != 1 {
		t.Errorf("unexpected report data: %+v", rep)
	}
}

func TestAllReportsHandler(t *testing.T) {
	h := report.NewHandler(&mockReportService{})

	ctx := context.WithValue(context.Background(), auth.RoleKey, "admin")
	req := httptest.NewRequest("GET", "/report/all", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	h.AllReports(w, req)

	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	var txs []transaction.Transaction
	if err := json.NewDecoder(res.Body).Decode(&txs); err != nil {
		t.Fatalf("decode err: %v", err)
	}
	if len(txs) != 2 || txs[0].ID != 1 {
		t.Errorf("unexpected txs: %+v", txs)
	}
}

func TestAdminReportHandler(t *testing.T) {
	h := report.NewHandler(&mockReportService{})

	ctx := context.WithValue(context.Background(), auth.RoleKey, "admin")
	req := httptest.NewRequest("GET", "/report/summary", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	h.AdminReport(w, req)

	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	var rep report.Report
	if err := json.NewDecoder(res.Body).Decode(&rep); err != nil {
		t.Fatalf("decode err: %v", err)
	}
	if rep.TotalCount != 5 || rep.SuccessCount != 3 || rep.FailedCount != 2 {
		t.Errorf("unexpected admin summary: %+v", rep)
	}
}
