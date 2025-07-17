package cache

import (
	"testing"
)

func TestSetAndGetTransaction(t *testing.T) {
	manager := GetManager()

	id := 12345
	data := "test-transaction-data"

	err := manager.SetTransaction(id, data)
	if err != nil {
		t.Fatalf("Failed to set transaction: %v", err)
	}

	result, err := manager.GetTransaction(id)
	if err != nil {
		t.Fatalf("Failed to get transaction: %v", err)
	}

	if result != data {
		t.Errorf("Expected %s, got %s", data, result)
	}
}
