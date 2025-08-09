package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	RoleKey   = contextKey("role")
	UserIDKey = contextKey("user_id")
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := ParseJWT(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		role, _ := claims["role"].(string)
		userIDFloat, _ := claims["user_id"].(float64)
		userID := int(userIDFloat)

		ctx := context.WithValue(r.Context(), RoleKey, role)
		ctx = context.WithValue(ctx, UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RoleRequired(required string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, _ := r.Context().Value(RoleKey).(string)
		if role != required {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RoleFromCtx(r *http.Request) string {
	role, _ := r.Context().Value(RoleKey).(string)
	return role
}
func UserIDFromCtx(r *http.Request) int {
	id, _ := r.Context().Value(UserIDKey).(int)
	return id
}
