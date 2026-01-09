package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"receiptScan-backend/internal/domain/receipt"
	"receiptScan-backend/internal/usecase/receipts"
)

type ReceiptsHandler struct {
	usecase receipts.UseCase
}

func NewReceiptsHandler(u receipts.UseCase) *ReceiptsHandler {
	return &ReceiptsHandler{usecase: u}
}

type receiptsResponse struct {
	Message string `json:"message"`
}

func (h *ReceiptsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req receipt.FormattedReceipt
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := h.usecase.Execute(r.Context(), receipts.Input{Receipt: req}); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, receipts.ErrInvalidUserName) {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(receiptsResponse{Message: "saved"})
}
