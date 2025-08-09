package transaction

import (
	"database/sql"
)

type Repository interface {
	Create(t *Transaction) error
	FindByID(id int) (*Transaction, error)
	ListByUser(userID int) ([]Transaction, error)
	AddTransaction(tx *Transaction) error
	DeleteTransaction(id int) error
	ListAll() ([]Transaction, error)
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
	row := r.db.QueryRow("SELECT id, user_id, amount, status, created_at FROM transactions WHERE id = ?", id)
	var tx Transaction
	err := row.Scan(&tx.ID, &tx.UserID, &tx.Amount, &tx.Status, &tx.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *DBRepo) ListByUser(userID int) ([]Transaction, error) {
	rows, err := r.db.Query("SELECT id, user_id, amount, status, created_at FROM transactions WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var txs []Transaction
	for rows.Next() {
		var tx Transaction
		err := rows.Scan(&tx.ID, &tx.UserID, &tx.Amount, &tx.Status, &tx.CreatedAt)
		if err != nil {
			continue
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

func (r *DBRepo) AddTransaction(tx *Transaction) error {
	result, err := r.db.Exec(
		"INSERT INTO transactions (user_id, amount, status) VALUES (?, ?, ?)",
		tx.UserID, tx.Amount, tx.Status,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	tx.ID = int(id)

	row := r.db.QueryRow("SELECT created_at FROM transactions WHERE id = ?", tx.ID)
	if err := row.Scan(&tx.CreatedAt); err != nil {
		return nil
	}
	return nil
}

func (r *DBRepo) DeleteTransaction(id int) error {
	_, err := r.db.Exec("DELETE FROM transactions WHERE id = ?", id)
	return err
}

func (r *DBRepo) ListAll() ([]Transaction, error) {
	rows, err := r.db.Query("SELECT id,user_id,amount,status,created_at FROM transactions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []Transaction
	for rows.Next() {
		var tx Transaction
		if err := rows.Scan(&tx.ID, &tx.UserID, &tx.Amount, &tx.Status, &tx.CreatedAt); err == nil {
			txs = append(txs, tx)
		}
	}
	return txs, nil
}
