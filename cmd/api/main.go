package main

import (
	"context"
	"log"
	"net/http"
	"os"

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

	// infra: Vision クライアント
	ocrService, err := vision.NewClient(ctx)
	if err != nil {
		log.Fatal("failed to initialize Vision client:", err)
	}

	// usecase
	ocrUsecase := ocr.NewUseCase(ocrService)

	// interface (HTTP handler)
	ocrHandler := iface.NewOcrHandler(ocrUsecase)

	http.Handle("/api/ocr", ocrHandler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
