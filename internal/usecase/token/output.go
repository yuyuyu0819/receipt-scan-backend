package token

import "time"

// CreateOutput は新規リフレッシュトークン発行の出力です。
type CreateOutput struct {
	Token     string
	ExpiresAt time.Time
}

// RefreshOutput はアクセストークン再発行の出力です。
type RefreshOutput struct {
	UserID int64
}
