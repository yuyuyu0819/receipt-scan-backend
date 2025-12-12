package ocr

import (
        "context"

        "receiptScan-backend/internal/domain/receipt"
)

// Formatter は OCR の生テキストを構造化されたレシートデータに整形します。
type Formatter interface {
        Format(ctx context.Context, rawText string) (receipt.FormattedReceipt, error)
}
