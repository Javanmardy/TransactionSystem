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

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	usr := h.userService.GetUserByUsername(req.Username)
	if usr == nil || usr.Password != req.Password {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}
	tokenString, err := GenerateJWT(usr.ID, usr.Username, usr.Role)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not generate token"})
		return
	}
	writeJSON(w, http.StatusOK, LoginResponse{Token: tokenString})
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	existingUser := h.userService.GetUserByUsername(req.Username)
	if existingUser != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "username already exists"})
		return
	}
	newUser := &user.User{
		Username: req.Username,
		Password: req.Password,
		Role:     "user",
		Email:    req.Email,
	}
	if err := h.userService.AddUser(newUser); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "user created"})
}
