package report

import (
	"TransactionSystem/internal/auth"
	"encoding/json"
	"net/http"
)

type Handler struct {
	svc Service
}

func NewHandler(s Service) *Handler {
	return &Handler{svc: s}
}

func (h *Handler) UserReport(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rep := h.svc.UserReport(userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rep)
}

func (h *Handler) AllReports(w http.ResponseWriter, r *http.Request) {
	role, _ := r.Context().Value(auth.RoleKey).(string)
	if role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	txs, err := h.svc.AllTransactions()
	if err != nil {
		http.Error(w, "failed to fetch transactions", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(txs)
}
func (h *Handler) AdminReport(w http.ResponseWriter, r *http.Request) {
	role, _ := r.Context().Value(auth.RoleKey).(string)
	if role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	rep := h.svc.AllReport()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rep)
}
