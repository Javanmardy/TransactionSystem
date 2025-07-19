package user_test

import (
	"TransactionSystem/internal/user"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockUserService struct{}

func (m *mockUserService) GetUserByID(id int) *user.User                { return nil }
func (m *mockUserService) GetUserByUsername(username string) *user.User { return nil }
func (m *mockUserService) AddUser(u *user.User) error                   { return nil }
func (m *mockUserService) ListAllUsers() ([]*user.User, error) {
	return []*user.User{
		{ID: 1, Username: "u1", Password: "p", Role: "user", Email: "u1@e.com"},
		{ID: 2, Username: "u2", Password: "p", Role: "admin", Email: "u2@e.com"},
	}, nil
}

func TestListUsersHandler(t *testing.T) {
	h := user.NewHandler(&mockUserService{})

	req := httptest.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()
	h.ListUsers(w, req)

	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", res.StatusCode)
	}
	var users []*user.User
	if err := json.NewDecoder(res.Body).Decode(&users); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(users) != 2 || users[0].Username != "u1" {
		t.Errorf("unexpected users result: %+v", users)
	}
}
