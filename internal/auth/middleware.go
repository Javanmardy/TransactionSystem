package auth

import (
	"context"
	"fmt"
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
		fmt.Println("[AuthMiddleware] Authorization header:", authHeader)

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			fmt.Println("[AuthMiddleware] Missing or invalid Authorization header")
			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		fmt.Println("[AuthMiddleware] Token string:", tokenString)

		claims, err := ParseJWT(tokenString)
		if err != nil {
			fmt.Println("[AuthMiddleware] Invalid or expired token:", err)
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		role, _ := claims["role"].(string)
		username, _ := claims["username"].(string)
		userIDFloat, _ := claims["user_id"].(float64)
		userID := int(userIDFloat)
		fmt.Println("[AuthMiddleware] User role from token:", role)
		fmt.Println("[AuthMiddleware] Username from token:", username)
		fmt.Println("[AuthMiddleware] UserID from token:", userID)

		ctx := context.WithValue(r.Context(), RoleKey, role)
		ctx = context.WithValue(ctx, UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
