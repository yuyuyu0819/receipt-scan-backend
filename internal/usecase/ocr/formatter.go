package ocr

import (
	"context"
	"errors"

	"receiptScan-backend/internal/domain/receipt"
)

// Formatter は OCR の生テキストを構造化されたレシートデータに整形します。
type Formatter interface {
	Format(ctx context.Context, rawText string) (receipt.FormattedReceipt, error)
}

// ErrInsufficientQuota は LLM 側の利用上限超過時に利用します。
var (
	ErrInsufficientQuota     = errors.New("formatter insufficient quota")
	ErrContextLengthExceeded = errors.New("formatter context length exceeded")
)
