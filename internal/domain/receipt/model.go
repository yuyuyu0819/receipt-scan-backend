package receipt

// OCR の結果（
type OcrResult struct {
	RawText string
	// 今後: ParsedItems []LineItem, TotalAmount, ShopName, PurchaseDate...
}
