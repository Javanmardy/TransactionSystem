package batch

import (
	"TransactionSystem/internal/transaction"
	"errors"
	"reflect"
	"testing"
)

type mockTxService struct {
	addErr error
}

func (m *mockTxService) GetTransactionByID(id int) *transaction.Transaction        { return nil }
func (m *mockTxService) ListUserTransactions(userID int) []transaction.Transaction { return nil }
func (m *mockTxService) AddTransaction(tx *transaction.Transaction) error          { return m.addErr }
func (m *mockTxService) AllTransactions() ([]transaction.Transaction, error)       { return nil, nil }
func (m *mockTxService) TransferFunds(fromUserID, toUserID int, amount float64, status string) error {
	return nil
}

func TestBatchProcessor_Process_AllScenarios(t *testing.T) {
	tests := []struct {
		name     string
		txs      []transaction.Transaction
		addErr   error
		wantStat []string
	}{
		{
			name:     "valid tx",
			txs:      []transaction.Transaction{{UserID: 1, Amount: 100, Status: "success"}},
			wantStat: []string{"success"},
		},
		{
			name:     "invalid amount",
			txs:      []transaction.Transaction{{UserID: 1, Amount: 0, Status: "success"}},
			wantStat: []string{"failed"},
		},
		{
			name:     "invalid status",
			txs:      []transaction.Transaction{{UserID: 1, Amount: 10, Status: "oops"}},
			wantStat: []string{"failed"},
		},
		{
			name:     "invalid user id",
			txs:      []transaction.Transaction{{UserID: -1, Amount: 10, Status: "success"}},
			wantStat: []string{"failed"},
		},
		{
			name:     "missing user id after adjustments",
			txs:      []transaction.Transaction{{Amount: 10, Status: "success"}},
			wantStat: []string{"failed"},
		},
		{
			name:     "AddTransaction error",
			txs:      []transaction.Transaction{{UserID: 1, Amount: 10, Status: "success"}},
			addErr:   errors.New("db fail"),
			wantStat: []string{"failed"},
		},
	}

	strats := []ValidationStrategy{
		&PositiveAmountStrategy{},
		&ValidStatusStrategy{},
		&PositiveUserIDStrategy{},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTxService{addErr: tt.addErr}
			bp := NewBatchProcessor(svc, strats)

			got := bp.Process(tt.txs, 99)
			var gotStat []string
			for _, r := range got {
				gotStat = append(gotStat, r.Status)
			}
			if !reflect.DeepEqual(gotStat, tt.wantStat) {
				t.Errorf("statuses mismatch: want %v, got %v", tt.wantStat, gotStat)
			}
		})
	}
}
