package batch

import (
	"TransactionSystem/internal/transaction"
	"encoding/json"
	"net/http"
)

type Handler struct {
	processor *BatchProcessor
}

func NewHandler(processor *BatchProcessor) *Handler {
	return &Handler{processor: processor}
}

func (h *Handler) ProcessBatch(w http.ResponseWriter, r *http.Request) {
	role, _ := r.Context().Value("role").(string)
	userID, _ := r.Context().Value("userID").(int)

	if role != "admin" {
		http.Error(w, "Forbidden: Only admin can process batch", http.StatusForbidden)
		return
	}

	var txs []transaction.Transaction
	if err := json.NewDecoder(r.Body).Decode(&txs); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	results := h.processor.Process(txs, userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
