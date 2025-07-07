package transaction

import (
	"testing"
)

func TestListUserTransactions(t *testing.T) {
	mockRepo := NewMockRepo()
	service := NewService(mockRepo)

	transactions := service.ListUserTransactions(1)
	if len(transactions) != 2 {
		t.Errorf("Expected 2 transactions for user 1, got %d", len(transactions))
	}
}

func TestGetTransactionByID(t *testing.T) {
	mockRepo := NewMockRepo()
	service := NewService(mockRepo)

	tx := service.GetTransactionByID(1)
	if tx == nil {
		t.Errorf("Expected transaction with ID 1, got nil")
	} else if tx.ID != 1 {
		t.Errorf("Expected transaction ID 1, got %d", tx.ID)
	}
}

func TestGetTransactionByID_NotFound(t *testing.T) {
	mockRepo := NewMockRepo()
	service := NewService(mockRepo)

	tx := service.GetTransactionByID(999)
	if tx != nil {
		t.Errorf("Expected nil for non-existent transaction, got %+v", tx)
	}
}
