package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"receiptScan-backend/internal/infra/db"
	"receiptScan-backend/internal/infra/openai"
	"receiptScan-backend/internal/infra/vision"
	iface "receiptScan-backend/internal/interface/http"
	"receiptScan-backend/internal/usecase/ocr"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env の読み込みに失敗:", err)
	}

	log.Println("GOOGLE_APPLICATION_CREDENTIALS =", os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))

	ctx := context.Background()

	dbURL := buildDatabaseURL()
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

	// interface (HTTP handler)
	ocrHandler := iface.NewOcrHandler(ocrUsecase)

	// OCR エンドポイント
	http.Handle("/api/ocr", ocrHandler)

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

// buildDatabaseURL は環境変数から接続文字列を構築します。
// DATABASE_URL が設定されている場合はそれを優先し、なければ
// docker-compose で用意したデフォルト (receipt/receipt@localhost:5432) を使います。
// 任意で PGHOST / PGPORT / PGUSER / PGPASSWORD / PGDATABASE / PGSSLMODE を上書きできます。
func buildDatabaseURL() string {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}

	host := getenvDefault("PGHOST", "localhost")
	port := getenvDefault("PGPORT", "5432")
	user := getenvDefault("PGUSER", "receipt")
	password := getenvDefault("PGPASSWORD", "receipt")
	database := getenvDefault("PGDATABASE", "receipt")
	sslMode := getenvDefault("PGSSLMODE", "disable")

	// URL エンコードを意識して url.URL を組み立てる
	u := &url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%s", host, port),
		Path:   "/" + database,
		User:   url.UserPassword(user, password),
	}

	q := u.Query()
	q.Set("sslmode", sslMode)
	u.RawQuery = q.Encode()

	return u.String()
}

func getenvDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
