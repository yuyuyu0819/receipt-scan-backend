package receiptslist

import "receiptScan-backend/internal/domain/receipt"

// Output is returned after fetching receipts for a user.
type Output struct {
	Receipts []receipt.Receipt
}
