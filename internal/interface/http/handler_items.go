package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"receiptScan-backend/internal/domain/receipt"
	"receiptScan-backend/internal/usecase/items"
)

type ItemsHandler struct {
	usecase items.UseCase
}

func NewItemsHandler(u items.UseCase) *ItemsHandler {
	return &ItemsHandler{usecase: u}
}

type itemsRequest struct {
	ReceiptID int64 `json:"receiptId"`
}

type itemsResponse struct {
	Items []receipt.Item `json:"items"`
}

func (h *ItemsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req itemsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	in := items.Input{ReceiptID: req.ReceiptID}
	out, err := h.usecase.Execute(r.Context(), in)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, items.ErrInvalidReceiptID) {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(itemsResponse{Items: out.Items})
}
