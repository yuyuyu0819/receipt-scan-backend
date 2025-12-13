package items

import "receiptScan-backend/internal/domain/receipt"

// Output is returned after fetching items linked to a receipt.
type Output struct {
	Items []receipt.Item
}
