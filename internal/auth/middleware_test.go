package auth_test

import (
	"TransactionSystem/internal/auth"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })

	auth.AuthMiddleware(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	if called {
		t.Fatalf("next handler should not be called on 401")
	}
}

func TestAuthMiddleware_BadToken(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer not-a-valid-jwt")
	rr := httptest.NewRecorder()

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })

	auth.AuthMiddleware(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	if called {
		t.Fatalf("next handler should not be called on bad token")
	}
}

func TestAuthMiddleware_OK(t *testing.T) {
	token, _ := auth.GenerateJWT(100, "u", "admin")
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := r.Context().Value(auth.RoleKey)
		userID := r.Context().Value(auth.UserIDKey)
		if role != "admin" || userID != 100 {
			t.Errorf("context values invalid: role=%v userID=%v", role, userID)
		}
		called = true
	})

	auth.AuthMiddleware(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !called {
		t.Fatalf("next handler was not called")
	}
}

func TestRoleRequired_Forbidden(t *testing.T) {
	token, _ := auth.GenerateJWT(5, "u", "user")
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	called := false
	protected := auth.RoleRequired("admin", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	auth.AuthMiddleware(protected).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
	if called {
		t.Fatalf("protected handler should not be called for wrong role")
	}
}

func TestRoleRequired_OK(t *testing.T) {
	token, _ := auth.GenerateJWT(7, "adminuser", "admin")
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	called := false
	protected := auth.RoleRequired("admin", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	auth.AuthMiddleware(protected).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !called {
		t.Fatalf("protected handler was not called")
	}
}
