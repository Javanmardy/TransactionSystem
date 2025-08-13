package transaction_test

import (
	"TransactionSystem/internal/transaction"
	"errors"
	"reflect"
	"testing"
)

type mockRepo struct {
	lastTransfer struct {
		fromUserID int
		toUserID   int
		amount     float64
		status     string
		called     bool
	}
	transferErr error

	addedTxs []transaction.Transaction

	listByUserRet []transaction.Transaction
	listAllRet    []transaction.Transaction
	findByIDRet   *transaction.Transaction
}

func (m *mockRepo) Create(t *transaction.Transaction) error { return nil }

func (m *mockRepo) FindByID(id int) (*transaction.Transaction, error) {
	return m.findByIDRet, nil
}

func (m *mockRepo) ListByUser(userID int) ([]transaction.Transaction, error) {
	return m.listByUserRet, nil
}

func (m *mockRepo) AddTransaction(tx *transaction.Transaction) error {
	cp := *tx
	m.addedTxs = append(m.addedTxs, cp)
	return nil
}

func (m *mockRepo) DeleteTransaction(id int) error { return nil }

func (m *mockRepo) ListAll() ([]transaction.Transaction, error) {
	return m.listAllRet, nil
}

func (m *mockRepo) TransferFunds(fromUserID, toUserID int, amount float64, status string) error {
	m.lastTransfer.called = true
	m.lastTransfer.fromUserID = fromUserID
	m.lastTransfer.toUserID = toUserID
	m.lastTransfer.amount = amount
	m.lastTransfer.status = status
	return m.transferErr
}

func TestTransferFunds_AmountNotPositive_LogsFailedAndReturnsError(t *testing.T) {
	repo := &mockRepo{}
	svc := transaction.NewService(repo)

	err := svc.TransferFunds(10, 20, 0, "success")
	if err == nil {
		t.Fatalf("expected error for non-positive amount, got nil")
	}

	if repo.lastTransfer.called {
		t.Errorf("repo.TransferFunds should not be called when amount <= 0")
	}

	if len(repo.addedTxs) != 1 {
		t.Fatalf("expected 1 failed log transaction, got %d", len(repo.addedTxs))
	}
	ft := repo.addedTxs[0]
	if ft.UserID != 10 || ft.FromUserID != 10 || ft.ToUserID != 20 || ft.Amount != 0 || ft.Status != "failed" {
		t.Errorf("unexpected failed tx log: %+v", ft)
	}
}

func TestTransferFunds_RepoError_LogsFailedAndReturnsError(t *testing.T) {
	repo := &mockRepo{transferErr: errors.New("db failure")}
	svc := transaction.NewService(repo)

	err := svc.TransferFunds(7, 8, 50, "success")
	if err == nil {
		t.Fatalf("expected error when repo.TransferFunds fails, got nil")
	}

	if !repo.lastTransfer.called || repo.lastTransfer.fromUserID != 7 || repo.lastTransfer.toUserID != 8 ||
		repo.lastTransfer.amount != 50 || repo.lastTransfer.status != "success" {
		t.Errorf("unexpected transfer call: %+v", repo.lastTransfer)
	}

	if len(repo.addedTxs) != 1 {
		t.Fatalf("expected 1 failed log transaction, got %d", len(repo.addedTxs))
	}
	ft := repo.addedTxs[0]
	if ft.UserID != 7 || ft.FromUserID != 7 || ft.ToUserID != 8 || ft.Amount != 0 || ft.Status != "failed" {
		t.Errorf("unexpected failed tx log: %+v", ft)
	}
}

func TestListUserTransactions_ReturnsRepoData(t *testing.T) {
	want := []transaction.Transaction{
		{ID: 3, UserID: 2, Amount: 10, Status: "success"},
		{ID: 2, UserID: 2, Amount: 20, Status: "failed"},
	}
	repo := &mockRepo{listByUserRet: want}
	svc := transaction.NewService(repo)

	got := svc.ListUserTransactions(2)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ListUserTransactions mismatch.\nwant: %+v\ngot:  %+v", want, got)
	}
}

func TestAllTransactions_ReturnsRepoData(t *testing.T) {
	want := []transaction.Transaction{
		{ID: 11, UserID: 5, Amount: 100, Status: "success"},
		{ID: 12, UserID: 6, Amount: 60, Status: "failed"},
	}
	repo := &mockRepo{listAllRet: want}
	svc := transaction.NewService(repo)

	got, err := svc.AllTransactions()
	if err != nil {
		t.Fatalf("AllTransactions returned error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("AllTransactions mismatch.\nwant: %+v\ngot:  %+v", want, got)
	}
}
