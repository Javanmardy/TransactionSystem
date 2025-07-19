package user

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	svc Service
}

func NewHandler(s Service) *Handler { return &Handler{svc: s} }

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.ListAllUsers()
	if err != nil {
		http.Error(w, "failed to fetch users", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
