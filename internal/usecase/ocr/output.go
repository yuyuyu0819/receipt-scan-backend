package ocr

import "receiptScan-backend/internal/domain/receipt"

type Output struct {
	Result    receipt.OcrResult
	Formatted *receipt.FormattedReceipt
}
