package transaction

type Service interface {
	GetTransactionByID(id int) *Transaction
	ListUserTransactions(userID int) []Transaction
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetTransactionByID(id int) *Transaction {
	tx, _ := s.repo.FindByID(id)
	return tx
}

func (s *service) ListUserTransactions(userID int) []Transaction {
	txs, _ := s.repo.ListByUser(userID)
	return txs
}
