package transaction

import "database/sql"

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
	db *sql.DB
}

func NewDBRepo(db *sql.DB) *DBRepo {
	return &DBRepo{db: db}
}

func (r *DBRepo) Create(t *Transaction) error {
	_, err := r.db.Exec("INSERT INTO transactions(user_id, amount, status) VALUES (?, ?, ?)", t.UserID, t.Amount, t.Status)
	return err
}

func (r *DBRepo) FindByID(id int) (*Transaction, error) {
	row := r.db.QueryRow("SELECT id, user_id, amount, status FROM transactions WHERE id = ?", id)
	var tx Transaction
	err := row.Scan(&tx.ID, &tx.UserID, &tx.Amount, &tx.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *DBRepo) ListByUser(userID int) ([]Transaction, error) {
	rows, err := r.db.Query("SELECT id, user_id, amount, status FROM transactions WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var txs []Transaction
	for rows.Next() {
		var tx Transaction
		err := rows.Scan(&tx.ID, &tx.UserID, &tx.Amount, &tx.Status)
		if err != nil {
			continue
		}
		txs = append(txs, tx)
	}
	return txs, nil
}
