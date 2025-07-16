package batch

import (
	"TransactionSystem/internal/auth"
	"TransactionSystem/internal/transaction"
	"encoding/json"
	"log"
	"net/http"
)

type Handler struct {
	txService transaction.Service
}

func NewHandler(txService transaction.Service) *Handler {
	return &Handler{txService: txService}
}

type BatchRequest struct {
	Transactions []transaction.Transaction `json:"transactions"`
}

func (h *Handler) ProcessBatch(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized (userID missing in token)", http.StatusUnauthorized)
		return
	}

	var req BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid batch request", http.StatusBadRequest)
		return
	}

	var results []BatchResult
	for _, tx := range req.Transactions {
		if tx.Amount <= 0 {
			results = append(results, BatchResult{
				Status: "failed",
				Error:  "amount must be positive",
			})
			continue
		}
		if tx.Status != "success" && tx.Status != "failed" {
			results = append(results, BatchResult{
				Status: "failed",
				Error:  "status must be 'success' or 'failed'",
			})
			continue
		}
		tx.UserID = userID
		if err := h.txService.AddTransaction(&tx); err != nil {
			log.Printf("Error adding transaction: %v", err)
			results = append(results, BatchResult{
				Status: "failed",
				Error:  "failed to save transaction",
			})
			continue
		}
		results = append(results, BatchResult{
			ID:     tx.ID,
			Status: "success",
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
