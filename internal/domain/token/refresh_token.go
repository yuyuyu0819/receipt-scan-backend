package token

import "time"

// RefreshToken は永続化されたリフレッシュトークンを表します。
type RefreshToken struct {
	ID        int64
	UserID    int64
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}
