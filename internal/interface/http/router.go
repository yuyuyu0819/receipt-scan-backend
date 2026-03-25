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
	"receiptScan-backend/internal/usecase/token"
)

// NewMux sets up HTTP routes and returns a ServeMux.
// Handlers are constructed here to keep DI and server startup logic in cmd/api/main.go.
func NewMux(ocrUsecase ocr.UseCase, itemsUsecase items.UseCase, receiptsUsecase receipts.UseCase, receiptsListUsecase receiptslist.UseCase, loginUsecase login.UseCase, signupUsecase signup.UseCase, tokenUsecase token.UseCase) *stdhttp.ServeMux {
	mux := stdhttp.NewServeMux()

	// interface (HTTP handler)
	ocrHandler := NewOcrHandler(ocrUsecase)
	itemsHandler := NewItemsHandler(itemsUsecase)
	receiptsHandler := NewReceiptsHandler(receiptsUsecase)
	receiptsListHandler := NewReceiptsListHandler(receiptsListUsecase)
	loginHandler := NewLoginHandler(loginUsecase, tokenUsecase)
	signupHandler := NewSignupHandler(signupUsecase)
	tokenHandler := NewTokenHandler(tokenUsecase)

	// OCR エンドポイント（認証必須）
	mux.Handle("/api/ocr", LoggingMiddleware(AuthMiddleware(ocrHandler)))
	// ログインエンドポイント（レートリミット付き）
	mux.Handle("/api/login", LoggingMiddleware(LoginRateLimitMiddleware(loginHandler)))
	// ユーザー作成エンドポイント（レートリミット付き）
	mux.Handle("/api/user/register", LoggingMiddleware(LoginRateLimitMiddleware(signupHandler)))
	// トークン更新エンドポイント（refreshToken → 新 accessToken）
	mux.Handle("/api/token/refresh", LoggingMiddleware(stdhttp.HandlerFunc(tokenHandler.ServeRefresh)))
	// トークン失効エンドポイント（ログアウト時に refreshToken を無効化）
	mux.Handle("/api/token/revoke", LoggingMiddleware(stdhttp.HandlerFunc(tokenHandler.ServeRevoke)))
	// レシート ID から items を取得するエンドポイント
	mux.Handle("/api/receipts/items", LoggingMiddleware(AuthMiddleware(itemsHandler)))
	// GET: ユーザーのレシート一覧取得 / POST: レシート登録
	mux.Handle("/api/receipts", LoggingMiddleware(AuthMiddleware(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		switch r.Method {
		case stdhttp.MethodGet:
			receiptsListHandler.ServeHTTP(w, r)
		case stdhttp.MethodPost:
			receiptsHandler.ServeHTTP(w, r)
		default:
			stdhttp.Error(w, "method not allowed", stdhttp.StatusMethodNotAllowed)
		}
	}))))

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
