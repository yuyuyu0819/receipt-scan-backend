package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"receiptScan-backend/internal/infra/db"
	"receiptScan-backend/internal/infra/openai"
	"receiptScan-backend/internal/infra/recaptcha"
	"receiptScan-backend/internal/infra/vision"
	iface "receiptScan-backend/internal/interface/http"
	"receiptScan-backend/internal/usecase/items"
	"receiptScan-backend/internal/usecase/login"
	"receiptScan-backend/internal/usecase/ocr"
	"receiptScan-backend/internal/usecase/receipts"
	"receiptScan-backend/internal/usecase/receiptslist"
	"receiptScan-backend/internal/usecase/signup"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env の読み込みに失敗:", err)
	}

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
	userRepo, err := db.NewUserRepository(ctx, dbURL)
	if err != nil {
		log.Fatal("failed to initialize user database:", err)
	}
	defer userRepo.Close()

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
	ocrUsecase := ocr.NewUseCase(ocrService, formatter)
	itemsUsecase := items.NewUseCase(receiptRepo)
	receiptsUsecase := receipts.NewUseCase(receiptRepo)
	receiptsListUsecase := receiptslist.NewUseCase(receiptRepo)
	loginUsecase := login.NewUseCase(userRepo)
	verifier, err := recaptcha.NewGoogleVerifierFromEnv()
	if err != nil {
		log.Fatal("failed to initialize reCAPTCHA verifier:", err)
	}
	signupUsecase := signup.NewUseCase(userRepo, verifier)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := iface.NewMux(ocrUsecase, itemsUsecase, receiptsUsecase, receiptsListUsecase, loginUsecase, signupUsecase)
	handler := iface.CORSMiddleware(mux)
	log.Println("Listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
