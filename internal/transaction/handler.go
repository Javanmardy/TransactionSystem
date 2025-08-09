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
		Amount float64 `json:"amount"`
		Status string  `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tx := &Transaction{
		UserID: userID,
		Amount: req.Amount,
		Status: req.Status,
	}

	err := h.service.AddTransaction(tx)
	if err != nil {
		http.Error(w, "Failed to add transaction", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tx)
}

func (h *Handler) UserTransactions(w http.ResponseWriter, r *http.Request) {
}
