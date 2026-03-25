package token

import (
	"context"
	"errors"
	"time"
)

// ErrTokenNotFound はトークンが DB に存在しない場合に返されます。
var ErrTokenNotFound = errors.New("refresh token not found")

// Repository はリフレッシュトークンの永続化インターフェースです。
type Repository interface {
	Save(ctx context.Context, userID int64, token string, expiresAt time.Time) error
	FindByToken(ctx context.Context, token string) (RefreshToken, error)
	Delete(ctx context.Context, token string) error
}
