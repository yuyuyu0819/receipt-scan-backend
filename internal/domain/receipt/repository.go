package receipt

import "context"

// Repository はレシートを永続化するためのインターフェースです。
type Repository interface {
	Save(ctx context.Context, receipt FormattedReceipt) error
	GetItemsByReceiptID(ctx context.Context, receiptID int64) ([]Item, error)
	GetReceiptsByUserID(ctx context.Context, userID string) ([]Receipt, error)
}
