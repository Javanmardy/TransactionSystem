package auth_test

import (
	"TransactionSystem/internal/auth"
	"TransactionSystem/internal/user"
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestDB(t *testing.T) user.Service {
	t.Helper()
	db, err := sql.Open("mysql", "root:n61224n61224@tcp(localhost:3306)/transaction_db")
	if err != nil {
		t.Fatalf("db open error: %v", err)
	}
	_, _ = db.Exec("DELETE FROM users")
	return user.NewMySQLService(db)
}

type mockUserService struct {
	userByUsername *user.User
	addErr         error
}

func (m *mockUserService) GetUserByID(id int) *user.User                { return nil }
func (m *mockUserService) GetUserByUsername(username string) *user.User { return m.userByUsername }
func (m *mockUserService) AddUser(u *user.User) error                   { return m.addErr }
func (m *mockUserService) ListAllUsers() ([]*user.User, error)          { return nil, nil }

func TestRegisterAndLogin(t *testing.T) {
	svc := setupTestDB(t)
	h := auth.NewHandler(svc)

	regBody := map[string]string{
		"username": "testuser1",
		"password": "p123",
		"email":    "testuser1@example.com",
	}
	body, _ := json.Marshal(regBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.RegisterHandler(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 created, got %d", w.Code)
	}

	loginBody := map[string]string{
		"username": "testuser1",
		"password": "p123",
	}
	body, _ = json.Marshal(loginBody)
	req = httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	w = httptest.NewRecorder()
	h.LoginHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 ok on login, got %d", w.Code)
	}
	var resp struct{ Token string }
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode token: %v", err)
	}
	if resp.Token == "" {
		t.Errorf("empty token in response")
	}
}

func TestLoginHandler_BadJSON(t *testing.T) {
	h := auth.NewHandler(&mockUserService{})
	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString("{bad-json"))
	w := httptest.NewRecorder()

	h.LoginHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	h := auth.NewHandler(&mockUserService{userByUsername: nil})
	body := `{"username":"u","password":"p"}`
	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.LoginHandler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing user, got %d", w.Code)
	}

	h = auth.NewHandler(&mockUserService{userByUsername: &user.User{ID: 1, Username: "u", Password: "correct", Role: "user"}})
	req = httptest.NewRequest("POST", "/login", bytes.NewBufferString(body))
	w = httptest.NewRecorder()
	h.LoginHandler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for wrong password, got %d", w.Code)
	}
}

func TestRegisterHandler_BadJSON(t *testing.T) {
	h := auth.NewHandler(&mockUserService{})
	req := httptest.NewRequest("POST", "/register", bytes.NewBufferString("{bad-json"))
	w := httptest.NewRecorder()

	h.RegisterHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRegisterHandler_Conflict(t *testing.T) {
	h := auth.NewHandler(&mockUserService{
		userByUsername: &user.User{ID: 1, Username: "duplicated"},
	})
	body := `{"username":"duplicated","password":"x","email":"d@d.com"}`
	req := httptest.NewRequest("POST", "/register", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.RegisterHandler(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestRegisterHandler_AddUserError(t *testing.T) {
	h := auth.NewHandler(&mockUserService{
		userByUsername: nil,
		addErr:         errors.New("db error"),
	})
	body := `{"username":"newuser","password":"x","email":"n@n.com"}`
	req := httptest.NewRequest("POST", "/register", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.RegisterHandler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestGenerateAndParseJWT(t *testing.T) {
	token, err := auth.GenerateJWT(10, "testuser", "admin")
	if err != nil {
		t.Fatalf("GenerateJWT error: %v", err)
	}
	claims, err := auth.ParseJWT(token)
	if err != nil {
		t.Fatalf("ParseJWT error: %v", err)
	}
	if claims["user_id"].(float64) != 10 || claims["username"] != "testuser" || claims["role"] != "admin" {
		t.Errorf("unexpected claims: %+v", claims)
	}
}
