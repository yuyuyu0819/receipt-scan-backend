package receipt

import "context"

// Repository はレシートを永続化するためのインターフェースです。
type Repository interface {
	Save(ctx context.Context, receipt FormattedReceipt) error
}
