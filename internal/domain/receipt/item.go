package receipt

// Item represents a stored receipt item row.
type Item struct {
	ID        int64  `json:"id"`
	ReceiptID int64  `json:"receiptId"`
	Name      string `json:"name"`
	Price     int    `json:"price"`
}
