package http

import (
	"context"
	"encoding/json"
	"net/http"

	"receiptScan-backend/internal/domain/receipt"
	"receiptScan-backend/internal/usecase/ocr"
)

type OcrHandler struct {
	usecase ocr.UseCase
}

func NewOcrHandler(u ocr.UseCase) *OcrHandler {
	return &OcrHandler{usecase: u}
}

// リクエスト DTO（外部向け）
type ocrRequest struct {
	ImageBase64 string `json:"imageBase64"`
}

// レスポンス DTO（外部向け）
type ocrResponse struct {
	Text      string                    `json:"text"`
	Formatted *receipt.FormattedReceipt `json:"formatted,omitempty"`
}

func (h *OcrHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// メソッド制御
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req ocrRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	in := ocr.Input{
		ImageBase64: req.ImageBase64,
	}

	out, err := h.usecase.Execute(context.Background(), in)
	if err != nil {
		http.Error(w, "OCR error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	res := ocrResponse{Text: out.Result.RawText, Formatted: out.Formatted}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}
