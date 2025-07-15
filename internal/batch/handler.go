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

type Handler struct {
	processor *BatchProcessor
}

func NewHandler(service transaction.Service) *Handler {
	strategies := []ValidationStrategy{
		&PositiveAmountStrategy{},
		&ValidStatusStrategy{},
	}
	return &Handler{
		processor: NewBatchProcessor(service, strategies),
	}
}

func (h *Handler) ProcessBatch(w http.ResponseWriter, r *http.Request) {
	var req BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid batch request", http.StatusBadRequest)
		return
	}
	results := h.processor.Process(req.Transactions, req.UserID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
