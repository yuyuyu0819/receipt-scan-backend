package receipts

import "receiptScan-backend/internal/domain/receipt"

// Input is the DTO for saving a formatted receipt.
type Input struct {
	Receipt receipt.FormattedReceipt
}
