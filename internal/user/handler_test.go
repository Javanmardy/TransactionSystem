package user_test

import (
	"TransactionSystem/internal/user"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockUserServiceOK struct{}

func (m *mockUserServiceOK) GetUserByID(id int) *user.User                { return nil }
func (m *mockUserServiceOK) GetUserByUsername(username string) *user.User { return nil }
func (m *mockUserServiceOK) AddUser(u *user.User) error                   { return nil }
func (m *mockUserServiceOK) ListAllUsers() ([]*user.User, error) {
	return []*user.User{
		{ID: 1, Username: "u1", Password: "p", Role: "user", Email: "u1@e.com"},
		{ID: 2, Username: "u2", Password: "p", Role: "admin", Email: "u2@e.com"},
	}, nil
}

type mockUserServiceErr struct{}

func (m *mockUserServiceErr) GetUserByID(id int) *user.User                { return nil }
func (m *mockUserServiceErr) GetUserByUsername(username string) *user.User { return nil }
func (m *mockUserServiceErr) AddUser(u *user.User) error                   { return nil }
func (m *mockUserServiceErr) ListAllUsers() ([]*user.User, error)          { return nil, fmt.Errorf("db error") }

func TestListUsersHandler_OK(t *testing.T) {
	h := user.NewHandler(&mockUserServiceOK{})

	req := httptest.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()
	h.ListUsers(w, req)

	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
	var users []*user.User
	if err := json.NewDecoder(res.Body).Decode(&users); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(users) != 2 || users[0].Username != "u1" {
		t.Errorf("unexpected users result: %+v", users)
	}
}

func TestListUsersHandler_Error(t *testing.T) {
	h := user.NewHandler(&mockUserServiceErr{})

	req := httptest.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()
	h.ListUsers(w, req)

	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Expected 500, got %d", res.StatusCode)
	}
}
