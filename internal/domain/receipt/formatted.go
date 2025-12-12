package receipt

// FormattedReceipt は ChatGPT によって整形されたレシート情報を表します。
type FormattedReceipt struct {
	Store string        `json:"store"`
	Date  string        `json:"date"`
	Total int           `json:"total"`
	Items []ReceiptItem `json:"items"`
}

// ReceiptItem はレシートの品目を表します。
type ReceiptItem struct {
	Name  string `json:"name"`
	Price int    `json:"price"`
}
