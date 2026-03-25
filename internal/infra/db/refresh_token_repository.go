package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domaintoken "receiptScan-backend/internal/domain/token"
)

// RefreshTokenRepository はリフレッシュトークンの PostgreSQL 実装です。
type RefreshTokenRepository struct {
	pool *pgxpool.Pool
}

// NewRefreshTokenRepository は接続プールを初期化してリポジトリを返します。
func NewRefreshTokenRepository(ctx context.Context, dsn string) (*RefreshTokenRepository, error) {
	if dsn == "" {
		return nil, errors.New("dsn is empty")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &RefreshTokenRepository{pool: pool}, nil
}

// Close は内部の接続プールをクローズします。
func (r *RefreshTokenRepository) Close() {
	r.pool.Close()
}

// Save は新しいリフレッシュトークンを保存します。
func (r *RefreshTokenRepository) Save(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO refresh_tokens (user_id, token, expires_at)
VALUES ($1, $2, $3)
`, userID, token, expiresAt)
	if err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}
	return nil
}

// FindByToken はトークン文字列でレコードを取得します。
func (r *RefreshTokenRepository) FindByToken(ctx context.Context, token string) (domaintoken.RefreshToken, error) {
	var rt domaintoken.RefreshToken
	err := r.pool.QueryRow(ctx, `
SELECT id, user_id, token, expires_at, created_at
FROM refresh_tokens
WHERE token = $1
`, token).Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domaintoken.RefreshToken{}, domaintoken.ErrTokenNotFound
		}
		return domaintoken.RefreshToken{}, fmt.Errorf("query refresh token: %w", err)
	}
	return rt, nil
}

// Delete はトークンを削除して無効化します。
func (r *RefreshTokenRepository) Delete(ctx context.Context, token string) error {
	_, err := r.pool.Exec(ctx, `
DELETE FROM refresh_tokens WHERE token = $1
`, token)
	if err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}
	return nil
}

var _ domaintoken.Repository = (*RefreshTokenRepository)(nil)
