package auth_test

import (
	"TransactionSystem/internal/auth"
	"TransactionSystem/internal/user"
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestDB(t *testing.T) user.Service {
	db, err := sql.Open("mysql", "root:n61224n61224@tcp(localhost:3306)/transaction_db")
	if err != nil {
		t.Fatalf("db open error: %v", err)
	}
	db.Exec("DELETE FROM users")
	return user.NewMySQLService(db)
}

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
func TestAuthMiddleware_OK(t *testing.T) {
	token, _ := auth.GenerateJWT(100, "u", "admin")
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := r.Context().Value(auth.RoleKey)
		userID := r.Context().Value(auth.UserIDKey)
		if role != "admin" || userID != 100 {
			t.Errorf("context values invalid: role=%v userID=%v", role, userID)
		}
		called = true
	})
	auth.AuthMiddleware(handler).ServeHTTP(rr, req)
	if !called {
		t.Fatal("handler was not called")
	}
}
