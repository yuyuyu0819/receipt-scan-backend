package ocr

import (
	"context"
	"errors"

	"receiptScan-backend/internal/domain/receipt"
)

// Vision や他の OCR サービスの抽象化インターフェース
type OCRService interface {
	DetectText(ctx context.Context, imageBase64 string) (receipt.OcrResult, error)
}

// ユースケースのインターフェース
type UseCase interface {
	Execute(ctx context.Context, in Input) (Output, error)
}

// 実装
type interactor struct {
	ocrService OCRService
}

func NewUseCase(ocrService OCRService) UseCase {
	return &interactor{
		ocrService: ocrService,
	}
}

func (i *interactor) Execute(ctx context.Context, in Input) (Output, error) {
	if in.ImageBase64 == "" {
		return Output{}, errors.New("imageBase64 is empty")
	}

	result, err := i.ocrService.DetectText(ctx, in.ImageBase64)
	if err != nil {
		return Output{}, err
	}

	return Output{Result: result}, nil
}
