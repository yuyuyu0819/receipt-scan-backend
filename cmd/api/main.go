package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"receiptScan-backend/internal/infra/db"
	"receiptScan-backend/internal/infra/openai"
	"receiptScan-backend/internal/infra/vision"
	iface "receiptScan-backend/internal/interface/http"
	"receiptScan-backend/internal/usecase/items"
	"receiptScan-backend/internal/usecase/ocr"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env の読み込みに失敗:", err)
	}

	log.Println("GOOGLE_APPLICATION_CREDENTIALS =", os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))

	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// docker-compose (postgres:16-alpine) のデフォルト設定
		dbURL = "postgres://receipt:receipt@localhost:5432/receipt?sslmode=disable"
	}
	receiptRepo, err := db.NewReceiptRepository(ctx, dbURL)
	if err != nil {
		log.Fatal("failed to initialize database:", err)
	}
	defer receiptRepo.Close()

	// infra: Vision クライアント
	ocrService, err := vision.NewClient(ctx)
	if err != nil {
		log.Fatal("failed to initialize Vision client:", err)
	}

	formatter, err := openai.NewFormatterClient()
	if err != nil {
		log.Fatal("failed to initialize OpenAI formatter:", err)
	}

	// usecase
	ocrUsecase := ocr.NewUseCase(ocrService, formatter, receiptRepo)
	itemsUsecase := items.NewUseCase(receiptRepo)

	// interface (HTTP handler)
	ocrHandler := iface.NewOcrHandler(ocrUsecase)
	itemsHandler := iface.NewItemsHandler(itemsUsecase)

	// OCR エンドポイント
	http.Handle("/api/ocr", ocrHandler)
	// レシート ID から items を取得するエンドポイント
	http.Handle("/api/receipts/items", itemsHandler)

	// health エンドポイント（ここにもログを追加）
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Println("[Health] path =", r.URL.Path, "method =", r.Method)
		w.Write([]byte("ok"))
	})

	// どのハンドラにもマッチしなかったときのフォールバック
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("[Fallback] path =", r.URL.Path, "method =", r.Method)
		http.NotFound(w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
