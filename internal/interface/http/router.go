package http

import (
	"log"
	stdhttp "net/http"

	"receiptScan-backend/internal/usecase/items"
	"receiptScan-backend/internal/usecase/ocr"
)

// NewMux sets up HTTP routes and returns a ServeMux.
// Handlers are constructed here to keep DI and server startup logic in cmd/api/main.go.
func NewMux(ocrUsecase ocr.UseCase, itemsUsecase items.UseCase) *stdhttp.ServeMux {
	mux := stdhttp.NewServeMux()

	// interface (HTTP handler)
	ocrHandler := NewOcrHandler(ocrUsecase)
	itemsHandler := NewItemsHandler(itemsUsecase)

	// OCR エンドポイント
	mux.Handle("/api/ocr", ocrHandler)
	// レシート ID から items を取得するエンドポイント
	mux.Handle("/api/receipts/items", itemsHandler)

	// health エンドポイント（ここにもログを追加）
	mux.HandleFunc("/health", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		log.Println("[Health] path =", r.URL.Path, "method =", r.Method)
		_, _ = w.Write([]byte("ok"))
	})

	// どのハンドラにもマッチしなかったときのフォールバック
	mux.HandleFunc("/", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		log.Println("[Fallback] path =", r.URL.Path, "method =", r.Method)
		stdhttp.NotFound(w, r)
	})

	return mux
}
