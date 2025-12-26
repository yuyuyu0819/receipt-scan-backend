package http

import (
	"log"
	stdhttp "net/http"

	"receiptScan-backend/internal/usecase/items"
	"receiptScan-backend/internal/usecase/login"
	"receiptScan-backend/internal/usecase/ocr"
	"receiptScan-backend/internal/usecase/receipts"
	"receiptScan-backend/internal/usecase/receiptslist"
	"receiptScan-backend/internal/usecase/signup"
)

// NewMux sets up HTTP routes and returns a ServeMux.
// Handlers are constructed here to keep DI and server startup logic in cmd/api/main.go.
func NewMux(ocrUsecase ocr.UseCase, itemsUsecase items.UseCase, receiptsUsecase receipts.UseCase, receiptsListUsecase receiptslist.UseCase, loginUsecase login.UseCase, signupUsecase signup.UseCase) *stdhttp.ServeMux {
	mux := stdhttp.NewServeMux()

	// interface (HTTP handler)
	ocrHandler := NewOcrHandler(ocrUsecase)
	itemsHandler := NewItemsHandler(itemsUsecase)
	receiptsHandler := NewReceiptsHandler(receiptsUsecase)
	receiptsListHandler := NewReceiptsListHandler(receiptsListUsecase)
	loginHandler := NewLoginHandler(loginUsecase)
	signupHandler := NewSignupHandler(signupUsecase)
	activationHandler := NewActivationHandler()

	// OCR エンドポイント
	mux.Handle("/api/ocr", LoggingMiddleware(ocrHandler))
	// ログインエンドポイント
	mux.Handle("/api/login", LoggingMiddleware(loginHandler))
	// ユーザー作成エンドポイント
	mux.Handle("/api/users", LoggingMiddleware(signupHandler))
	// アカウント有効化ページ
	mux.Handle("/activate", LoggingMiddleware(activationHandler))
	// レシート ID から items を取得するエンドポイント
	mux.Handle("/api/receipts/items", LoggingMiddleware(itemsHandler))
	// レシート内容を登録するエンドポイント
	mux.Handle("/api/receipts", LoggingMiddleware(receiptsHandler))
	// ユーザー ID からレシートを取得するエンドポイント
	mux.Handle("/api/receipts/by-user", LoggingMiddleware(receiptsListHandler))

	// health エンドポイント（ここにもログを追加）
	mux.Handle("/health", LoggingMiddleware(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		log.Println("[Health] path =", r.URL.Path, "method =", r.Method)
		_, _ = w.Write([]byte("ok"))
	})))

	// どのハンドラにもマッチしなかったときのフォールバック
	mux.Handle("/", LoggingMiddleware(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		log.Println("[Fallback] path =", r.URL.Path, "method =", r.Method)
		stdhttp.NotFound(w, r)
	})))

	return mux
}
