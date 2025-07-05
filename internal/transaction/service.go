package transaction

type Transaction struct {
	ID     int
	Amount float64
	UserID int
	Status string
}

type Service interface {
	GetTransactionByID(id int) *Transaction
	ListUserTransactions(userID int) []Transaction
}

type mockService struct{}
type mysqlService struct{}

func NewService(mode string) Service {
	if mode == "mock" {
		return &mockService{}
	}
	return &mysqlService{}
}

// -------------------- MOCK --------------------

var mockTransactions = []Transaction{
	{ID: 1, Amount: 1000, UserID: 1, Status: "success"},
	{ID: 2, Amount: 500, UserID: 1, Status: "fail"},
	{ID: 3, Amount: 2000, UserID: 2, Status: "success"},
}

func (s *mockService) GetTransactionByID(id int) *Transaction {
	for _, t := range mockTransactions {
		if t.ID == id {
			return &t
		}
	}
	return nil
}

func (s *mockService) ListUserTransactions(userID int) []Transaction {
	var result []Transaction
	for _, t := range mockTransactions {
		if t.UserID == userID {
			result = append(result, t)
		}
	}
	return result
}

// -------------------- MySQL --------------------
func (s *mysqlService) GetTransactionByID(id int) *Transaction {
	return nil
}

func (s *mysqlService) ListUserTransactions(userID int) []Transaction {
	return nil
}
