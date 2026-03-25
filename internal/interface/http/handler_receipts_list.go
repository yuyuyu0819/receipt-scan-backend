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

type receiptsListResponse struct {
	Receipts []receipt.Receipt `json:"receipts"`
}

func (h *ReceiptsListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := userIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	out, err := h.usecase.Execute(r.Context(), receiptslist.Input{UserID: userID})
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
