package user_test

import (
	"TransactionSystem/internal/user"
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
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
	if err := svc.AddUser(u); err != nil {
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

func TestGetUserByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	svc := user.NewMySQLService(db)

	got := svc.GetUserByID(999999)
	if got != nil {
		t.Errorf("Expected nil, got %+v", got)
	}
}

func TestGetUserByUsername_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	svc := user.NewMySQLService(db)

	got := svc.GetUserByUsername("no_user_hopefully")
	if got != nil {
		t.Errorf("Expected nil, got %+v", got)
	}
}

func TestListAllUsers_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	svc := user.NewMySQLService(db)

	users, err := svc.ListAllUsers()
	if err != nil {
		t.Fatalf("ListAllUsers error: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("Expected empty list, got %d users", len(users))
	}
}

func TestAddUser_DBClosed_Error(t *testing.T) {
	db := setupTestDB(t)
	svc := user.NewMySQLService(db)
	db.Close()
	u := &user.User{Username: "x", Password: "y", Role: "user", Email: "x@y.z"}
	if err := svc.AddUser(u); err == nil {
		t.Fatalf("expected error when DB is closed, got nil")
	}
}
