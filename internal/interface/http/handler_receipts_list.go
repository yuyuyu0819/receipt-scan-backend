package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"receiptScan-backend/internal/domain/receipt"
	"receiptScan-backend/internal/usecase/receiptslist"
)

type ReceiptsListHandler struct {
	usecase receiptslist.UseCase
}

func NewReceiptsListHandler(u receiptslist.UseCase) *ReceiptsListHandler {
	return &ReceiptsListHandler{usecase: u}
}

type receiptsListRequest struct {
	UserID int64 `json:"userId"`
}

type receiptsListResponse struct {
	Receipts []receipt.Receipt `json:"receipts"`
}

func (h *ReceiptsListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req receiptsListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	out, err := h.usecase.Execute(r.Context(), receiptslist.Input{UserID: req.UserID})
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, receiptslist.ErrInvalidUserID) {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(receiptsListResponse{Receipts: out.Receipts})
}
