package transaction

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func InitDB(user, password, host, dbname string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", user, password, host, dbname)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	schema := `
CREATE TABLE IF NOT EXISTS transactions (
	id INT AUTO_INCREMENT PRIMARY KEY,
	user_id INT,
	amount DOUBLE,
	status VARCHAR(20),
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`
	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}
	return db, nil
}
