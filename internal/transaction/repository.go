package transaction

import (
	"database/sql"
	"fmt"
)

type Repository interface {
	Create(t *Transaction) error
	FindByID(id int) (*Transaction, error)
	ListByUser(userID int) ([]Transaction, error)
	AddTransaction(tx *Transaction) error
	DeleteTransaction(id int) error
	ListAll() ([]Transaction, error)
	TransferFunds(fromUserID, toUserID int, amount float64, status string) error
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
	row := r.db.QueryRow(`
		SELECT id, user_id,
		       COALESCE(from_user_id,0),
		       COALESCE(to_user_id,0),
		       amount, status, created_at
		FROM transactions
		WHERE id = ?`, id)

	var tx Transaction
	err := row.Scan(&tx.ID, &tx.UserID, &tx.FromUserID, &tx.ToUserID, &tx.Amount, &tx.Status, &tx.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *DBRepo) ListByUser(userID int) ([]Transaction, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id,
		       COALESCE(from_user_id,0),
		       COALESCE(to_user_id,0),
		       amount, status, created_at
		FROM transactions
		WHERE user_id = ?
		ORDER BY id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []Transaction
	for rows.Next() {
		var tx Transaction
		if err := rows.Scan(&tx.ID, &tx.UserID, &tx.FromUserID, &tx.ToUserID, &tx.Amount, &tx.Status, &tx.CreatedAt); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

func (r *DBRepo) AddTransaction(txn *Transaction) error {
	var from sql.NullInt64
	if txn.FromUserID != 0 {
		from = sql.NullInt64{Int64: int64(txn.FromUserID), Valid: true}
	}
	var to sql.NullInt64
	if txn.ToUserID != 0 {
		to = sql.NullInt64{Int64: int64(txn.ToUserID), Valid: true}
	}

	res, err := r.db.Exec(
		`INSERT INTO transactions (user_id, from_user_id, to_user_id, amount, status)
		 VALUES (?, ?, ?, ?, ?)`,
		txn.UserID, from, to, txn.Amount, txn.Status,
	)
	if err != nil {
		return err
	}

	if _, err := r.db.Exec(
		`UPDATE users SET balance = balance + ? WHERE id = ?`,
		txn.Amount, txn.UserID,
	); err != nil {
		return err
	}

	id64, _ := res.LastInsertId()
	txn.ID = int(id64)

	row := r.db.QueryRow(`SELECT created_at FROM transactions WHERE id = ?`, txn.ID)
	if err := row.Scan(&txn.CreatedAt); err != nil && err != sql.ErrNoRows {
		return err
	}

	return nil
}

func (r *DBRepo) DeleteTransaction(id int) error {
	_, err := r.db.Exec("DELETE FROM transactions WHERE id = ?", id)
	return err
}
func (r *DBRepo) TransferFunds(fromUserID, toUserID int, amount float64, status string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var fromBal float64
	err = tx.QueryRow("SELECT balance FROM users WHERE id = ? FOR UPDATE", fromUserID).Scan(&fromBal)
	if err == sql.ErrNoRows {
		return fmt.Errorf("sender not found")
	}
	if err != nil {
		return err
	}
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	if fromBal < amount {
		return fmt.Errorf("insufficient funds")
	}

	var toBal float64
	err = tx.QueryRow("SELECT balance FROM users WHERE id = ? FOR UPDATE", toUserID).Scan(&toBal)
	if err == sql.ErrNoRows {
		return fmt.Errorf("receiver not found")
	}
	if err != nil {
		return err
	}

	if _, err = tx.Exec(
		"UPDATE users SET balance = balance - ? WHERE id = ?", amount, fromUserID,
	); err != nil {
		return err
	}
	if _, err = tx.Exec(
		"UPDATE users SET balance = balance + ? WHERE id = ?", amount, toUserID,
	); err != nil {
		return err
	}

	if _, err = tx.Exec(
		"INSERT INTO transactions (user_id, from_user_id, to_user_id, amount, status) VALUES (?, ?, ?, ?, ?)",
		fromUserID, fromUserID, toUserID, -amount, status,
	); err != nil {
		return err
	}

	if _, err = tx.Exec(
		"INSERT INTO transactions (user_id, from_user_id, to_user_id, amount, status) VALUES (?, ?, ?, ?, ?)",
		toUserID, fromUserID, toUserID, amount, status,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *DBRepo) ListAll() ([]Transaction, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id,
		       COALESCE(from_user_id,0),
		       COALESCE(to_user_id,0),
		       amount, status, created_at
		FROM transactions
		WHERE amount > -1
		ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []Transaction
	for rows.Next() {
		var tx Transaction
		if err := rows.Scan(&tx.ID, &tx.UserID, &tx.FromUserID, &tx.ToUserID, &tx.Amount, &tx.Status, &tx.CreatedAt); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}
