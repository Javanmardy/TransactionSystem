package transaction

import (
	"TransactionSystem/internal/cache"
	"encoding/json"
	"fmt"
	"log"
	"time"
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
	TransferFunds(fromUserID, toUserID int, amount float64, status string) error
}

func (s *service) TransferFunds(fromUserID, toUserID int, amount float64, status string) error {
	if amount <= 0 {
		_ = s.repo.AddTransaction(&Transaction{
			UserID:     fromUserID,
			FromUserID: fromUserID,
			ToUserID:   toUserID,
			Amount:     0,
			Status:     "failed",
		})
		return fmt.Errorf("amount must be positive")
	}

	if err := s.repo.TransferFunds(fromUserID, toUserID, amount, status); err != nil {
		_ = s.repo.AddTransaction(&Transaction{
			UserID:     fromUserID,
			FromUserID: fromUserID,
			ToUserID:   toUserID,
			Amount:     0,
			Status:     "failed",
		})
		return err
	}

	cm := cache.GetManager()
	now := time.Now()
	senderTx := &Transaction{
		UserID:     fromUserID,
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Amount:     -amount,
		Status:     status,
		CreatedAt:  now,
	}
	receiverTx := &Transaction{
		UserID:     toUserID,
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Amount:     amount,
		Status:     status,
		CreatedAt:  now,
	}
	b1, _ := json.Marshal(senderTx)
	b2, _ := json.Marshal(receiverTx)
	log.Printf("[debug] PushRecent from=%d to=%d", fromUserID, toUserID)
	if err := cm.PushRecent(fromUserID, string(b1), 20); err != nil {
		log.Printf("[error] pushRecent from-user %d: %v", fromUserID, err)
	}
	if err := cm.PushRecent(toUserID, string(b2), 20); err != nil {
		log.Printf("[error] pushRecent to-user %d: %v", toUserID, err)
	}
	return nil
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
