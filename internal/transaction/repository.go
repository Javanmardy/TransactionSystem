package transaction

type Transaction struct {
	ID     int
	UserID int
	Amount float64
	Status string
}

type Repository interface {
	Create(t *Transaction) error
	FindByID(id int) (*Transaction, error)
	ListByUser(userID int) ([]Transaction, error)
}

type MockRepo struct {
	data []Transaction
}

func NewMockRepo() *MockRepo {
	return &MockRepo{
		data: []Transaction{
			{ID: 1, UserID: 1, Amount: 1000, Status: "success"},
			{ID: 2, UserID: 1, Amount: 200, Status: "failed"},
			{ID: 3, UserID: 2, Amount: 300, Status: "success"},
		},
	}
}

func (r *MockRepo) Create(t *Transaction) error {
	t.ID = len(r.data) + 1
	r.data = append(r.data, *t)
	return nil
}

func (r *MockRepo) FindByID(id int) (*Transaction, error) {
	for _, v := range r.data {
		if v.ID == id {
			return &v, nil
		}
	}
	return nil, nil
}

func (r *MockRepo) ListByUser(userID int) ([]Transaction, error) {
	var result []Transaction
	for _, v := range r.data {
		if v.UserID == userID {
			result = append(result, v)
		}
	}
	return result, nil
}

type DBRepo struct {
	// db *sql.DB
}

func NewDBRepo() *DBRepo {
	return &DBRepo{}
}

func (r *DBRepo) Create(t *Transaction) error {
	return nil
}

func (r *DBRepo) FindByID(id int) (*Transaction, error) {
	return nil, nil
}

func (r *DBRepo) ListByUser(userID int) ([]Transaction, error) {
	return nil, nil
}
