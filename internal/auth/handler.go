package auth

import (
	"TransactionSystem/internal/user"
	"encoding/json"
	"net/http"
)

type Handler struct {
	userService user.Service
}

func NewHandler(userService user.Service) *Handler {
	return &Handler{userService: userService}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type LoginResponse struct {
	Token string `json:"token"`
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	usr := h.userService.GetUserByUsername(req.Username)
	if usr == nil || usr.Password != req.Password {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	tokenString, err := GenerateJWT(usr.ID, usr.Username, usr.Role)
	if err != nil {
		http.Error(w, "could not generate token", http.StatusInternalServerError)
		return
	}

	resp := LoginResponse{Token: tokenString}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
