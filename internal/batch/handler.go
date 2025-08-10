package batch

import (
	"TransactionSystem/internal/auth"
	"TransactionSystem/internal/transaction"
	"encoding/json"
	"net/http"
)

type BatchRequest struct {
	Transactions []transaction.Transaction `json:"transactions"`
}

type Handler struct {
	processor *BatchProcessor
}

func NewHandler(txSvc transaction.Service) *Handler {
	strats := []ValidationStrategy{
		&PositiveAmountStrategy{},
		&ValidStatusStrategy{},
		&PositiveUserIDStrategy{},
	}
	return &Handler{
		processor: NewBatchProcessor(txSvc, strats),
	}
}

func (h *Handler) ProcessBatch(w http.ResponseWriter, r *http.Request) {
	role, _ := r.Context().Value(auth.RoleKey).(string)
	if role != "admin" {
		http.Error(w, "Forbidden: Only admin can process batch", http.StatusForbidden)
		return
	}

	actorID, _ := r.Context().Value(auth.UserIDKey).(int)

	var req BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	results := h.processor.Process(req.Transactions, actorID)
	json.NewEncoder(w).Encode(results)
}
