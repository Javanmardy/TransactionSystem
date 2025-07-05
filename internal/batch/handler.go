package batch

import (
	"TransactionSystem/internal/transaction"
	"encoding/json"
	"net/http"
)

type BatchRequest struct {
	UserID       int                       `json:"user_id"`
	Transactions []transaction.Transaction `json:"transactions"`
}

type BatchResult struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type Handler struct {
	txService transaction.Service
}

func NewHandler(txSvc transaction.Service) *Handler {
	return &Handler{txService: txSvc}
}

func (h *Handler) ProcessBatch(w http.ResponseWriter, r *http.Request) {
	var req BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	var results []BatchResult
	for _, tx := range req.Transactions {
		if tx.Amount <= 0 {
			results = append(results, BatchResult{
				ID:     tx.ID,
				Status: "failed",
				Error:  "amount must be positive",
			})
			continue
		}
		tx.UserID = req.UserID
		results = append(results, BatchResult{
			ID:     tx.ID,
			Status: "success",
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
