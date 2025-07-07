package auth

import (
	"encoding/json"
	"net/http"
)

var users = map[string]struct {
	Password string
	Role     string
}{
	"amirhossein": {"adminpass", "admin"},
	"sara":        {"test", "user"},
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type LoginResponse struct {
	Token string `json:"token"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, ok := users[req.Username]
	if !ok || user.Password != req.Password {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	tokenString, err := GenerateJWT(req.Username, user.Role)
	if err != nil {
		http.Error(w, "could not generate token", http.StatusInternalServerError)
		return
	}

	resp := LoginResponse{Token: tokenString}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
