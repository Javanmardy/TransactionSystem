package transaction

import (
	"TransactionSystem/internal/auth"
	"encoding/json"
	"net/http"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) ListUserTransactions(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized (userID missing in token)", http.StatusUnauthorized)
		return
	}

	transactions := h.service.ListUserTransactions(userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

func (h *Handler) AddTransactionHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ToUserID int     `json:"to_user_id"`
		Amount   float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.ToUserID == 0 || req.ToUserID == userID {
		http.Error(w, "Invalid recipient", http.StatusBadRequest)
		return
	}
	if req.Amount <= 0 {
		http.Error(w, "amount must be positive", http.StatusBadRequest)
		return
	}

	if err := h.service.TransferFunds(userID, req.ToUserID, req.Amount, "success"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message":   "transfer success",
		"from_user": userID,
		"to_user":   req.ToUserID,
		"amount":    req.Amount,
		"status":    "success",
	})
}

func (h *Handler) UserTransactions(w http.ResponseWriter, r *http.Request) {
}
