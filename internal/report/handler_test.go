package report_test

import (
	"TransactionSystem/internal/auth"
	"TransactionSystem/internal/report"
	"TransactionSystem/internal/transaction"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockReportServiceOK struct{}

func (m *mockReportServiceOK) UserReport(userID int) report.Report {
	return report.Report{TotalCount: 2, SuccessCount: 1, FailedCount: 1}
}
func (m *mockReportServiceOK) AllTransactions() ([]transaction.Transaction, error) {
	return []transaction.Transaction{{ID: 1}, {ID: 2}}, nil
}
func (m *mockReportServiceOK) AllReport() report.Report {
	return report.Report{TotalCount: 5, SuccessCount: 3, FailedCount: 2}
}

type mockReportServiceErr struct{ mockReportServiceOK }

func (m *mockReportServiceErr) AllTransactions() ([]transaction.Transaction, error) {
	return nil, errors.New("db error")
}

func TestUserReportHandler_OK(t *testing.T) {
	h := report.NewHandler(&mockReportServiceOK{})
	ctx := context.WithValue(context.Background(), auth.UserIDKey, 7)
	req := httptest.NewRequest("GET", "/report", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	h.UserReport(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var rep report.Report
	if err := json.NewDecoder(w.Body).Decode(&rep); err != nil {
		t.Fatalf("decode err: %v", err)
	}
	if rep.TotalCount != 2 {
		t.Errorf("unexpected data: %+v", rep)
	}
}

func TestUserReportHandler_Unauthorized(t *testing.T) {
	h := report.NewHandler(&mockReportServiceOK{})
	req := httptest.NewRequest("GET", "/report", nil)
	w := httptest.NewRecorder()
	h.UserReport(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAllReportsHandler_OK(t *testing.T) {
	h := report.NewHandler(&mockReportServiceOK{})
	ctx := context.WithValue(context.Background(), auth.RoleKey, "admin")
	req := httptest.NewRequest("GET", "/report/all", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	h.AllReports(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAllReportsHandler_Forbidden(t *testing.T) {
	h := report.NewHandler(&mockReportServiceOK{})
	ctx := context.WithValue(context.Background(), auth.RoleKey, "user")
	req := httptest.NewRequest("GET", "/report/all", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	h.AllReports(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestAllReportsHandler_ErrorFromService(t *testing.T) {
	h := report.NewHandler(&mockReportServiceErr{})
	ctx := context.WithValue(context.Background(), auth.RoleKey, "admin")
	req := httptest.NewRequest("GET", "/report/all", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	h.AllReports(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestAdminReportHandler_OK(t *testing.T) {
	h := report.NewHandler(&mockReportServiceOK{})
	ctx := context.WithValue(context.Background(), auth.RoleKey, "admin")
	req := httptest.NewRequest("GET", "/report/summary", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	h.AdminReport(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAdminReportHandler_Forbidden(t *testing.T) {
	h := report.NewHandler(&mockReportServiceOK{})
	ctx := context.WithValue(context.Background(), auth.RoleKey, "user")
	req := httptest.NewRequest("GET", "/report/summary", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	h.AdminReport(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}
