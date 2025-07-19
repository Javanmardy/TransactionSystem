package user_test

import (
	"TransactionSystem/internal/user"
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("mysql", "root:n61224n61224@tcp(localhost:3306)/transaction_db?parseTime=true")
	if err != nil {
		t.Fatalf("failed to connect db: %v", err)
	}
	_, _ = db.Exec("DELETE FROM users")
	return db
}

func TestAddUserAndGetUserByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	svc := user.NewMySQLService(db)

	u := &user.User{
		Username: "testuser",
		Password: "pass",
		Role:     "user",
		Email:    "t@t.com",
	}
	err := svc.AddUser(u)
	if err != nil {
		t.Fatalf("AddUser error: %v", err)
	}
	got := svc.GetUserByID(u.ID)
	if got == nil || got.Username != u.Username {
		t.Errorf("GetUserByID failed, want %s got %+v", u.Username, got)
	}
}

func TestGetUserByUsername(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	svc := user.NewMySQLService(db)

	u := &user.User{
		Username: "johndoe",
		Password: "abc123",
		Role:     "admin",
		Email:    "john@example.com",
	}
	_ = svc.AddUser(u)

	got := svc.GetUserByUsername("johndoe")
	if got == nil || got.Email != "john@example.com" {
		t.Errorf("GetUserByUsername failed, got %+v", got)
	}
}

func TestListAllUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	svc := user.NewMySQLService(db)

	_ = svc.AddUser(&user.User{Username: "u1", Password: "p1", Role: "user", Email: "a@a.com"})
	_ = svc.AddUser(&user.User{Username: "u2", Password: "p2", Role: "admin", Email: "b@b.com"})

	users, err := svc.ListAllUsers()
	if err != nil {
		t.Fatalf("ListAllUsers error: %v", err)
	}
	if len(users) < 2 {
		t.Errorf("expected at least 2 users, got %d", len(users))
	}
}
