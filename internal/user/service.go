package user

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type Service interface {
	GetUserByID(id int) *User
	GetUserByUsername(username string) *User
	AddUser(user *User) error
}

type mysqlService struct {
	db *sql.DB
}

func NewMySQLService(db *sql.DB) Service {
	return &mysqlService{db: db}
}

func (s *mysqlService) GetUserByID(id int) *User {
	user := &User{}
	err := s.db.QueryRow("SELECT id, username, password, role, email FROM users WHERE id=?", id).
		Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.Email)
	if err != nil {
		return nil
	}
	return user
}

func (s *mysqlService) GetUserByUsername(username string) *User {
	user := &User{}
	err := s.db.QueryRow("SELECT id, username, password, role, email FROM users WHERE username=?", username).
		Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.Email)
	if err != nil {
		return nil
	}
	return user
}

func (s *mysqlService) AddUser(user *User) error {
	res, err := s.db.Exec("INSERT INTO users (username, password, role, email) VALUES (?, ?, ?, ?)",
		user.Username, user.Password, user.Role, user.Email)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	user.ID = int(id)
	return nil
}
