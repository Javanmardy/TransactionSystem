package report

import (
	"TransactionSystem/internal/transaction"
	"reflect"
	"testing"
)

type mockTxService struct {
	userTxs []transaction.Transaction
	allTxs  []transaction.Transaction
}

func (m *mockTxService) GetTransactionByID(id int) *transaction.Transaction { return nil }
func (m *mockTxService) ListUserTransactions(userID int) []transaction.Transaction {
	return m.userTxs
}
func (m *mockTxService) AddTransaction(tx *transaction.Transaction) error { return nil }
func (m *mockTxService) AllTransactions() ([]transaction.Transaction, error) {
	return m.allTxs, nil
}
func (m *mockTxService) TransferFunds(fromUserID, toUserID int, amount float64, status string) error {
	return nil
}

func TestUserReport_Calculations(t *testing.T) {
	txSvc := &mockTxService{
		userTxs: []transaction.Transaction{
			{Amount: 100, Status: "success"},
			{Amount: 50, Status: "failed"},
		},
	}
	svc := NewService(txSvc)

	got := svc.UserReport(1)
	want := Report{
		TotalCount:    2,
		SuccessCount:  1,
		FailedCount:   1,
		TotalAmount:   150,
		SuccessAmount: 100,
		FailedAmount:  50,
		SuccessRate:   0.5,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("UserReport mismatch\nwant: %+v\ngot:  %+v", want, got)
	}
}

func TestAllReport_Calculations(t *testing.T) {
	txSvc := &mockTxService{
		allTxs: []transaction.Transaction{
			{Amount: 20, Status: "success"},
			{Amount: 30, Status: "failed"},
			{Amount: 50, Status: "success"},
		},
	}
	svc := NewService(txSvc)

	got := svc.AllReport()
	want := Report{
		TotalCount:    3,
		SuccessCount:  2,
		FailedCount:   1,
		TotalAmount:   100,
		SuccessAmount: 70,
		FailedAmount:  30,
		SuccessRate:   2.0 / 3.0,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("AllReport mismatch\nwant: %+v\ngot:  %+v", want, got)
	}
}
