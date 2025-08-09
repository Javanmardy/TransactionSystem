package transaction

import (
	"TransactionSystem/internal/cache"
	"encoding/json"
)

func (s *service) GetTransactionByID(id int) *Transaction {
	cm := cache.GetManager()
	if data, err := cm.GetTransaction(id); err == nil && data != "" {
		var tx Transaction
		if err := json.Unmarshal([]byte(data), &tx); err == nil {
			return &tx
		}
	}
	tx, _ := s.repo.FindByID(id)
	if tx != nil {
		data, _ := json.Marshal(tx)
		_ = cm.SetTransaction(tx.ID, string(data))
	}
	return tx
}

type Service interface {
	GetTransactionByID(id int) *Transaction
	ListUserTransactions(userID int) []Transaction
	AddTransaction(tx *Transaction) error
	AllTransactions() ([]Transaction, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) ListUserTransactions(userID int) []Transaction {
	txs, _ := s.repo.ListByUser(userID)
	return txs
}

func (s *service) AddTransaction(tx *Transaction) error {
	if err := s.repo.AddTransaction(tx); err != nil {
		return err
	}
	cm := cache.GetManager()
	b, _ := json.Marshal(tx)
	_ = cm.SetTransaction(tx.ID, string(b))
	_ = cm.PushRecent(tx.UserID, string(b), 20)
	return nil
}

func (s *service) AllTransactions() ([]Transaction, error) {
	return s.repo.ListAll()
}
