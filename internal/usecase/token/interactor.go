package token

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	domaintoken "receiptScan-backend/internal/domain/token"
)

const refreshTokenTTL = 30 * 24 * time.Hour

// ErrTokenExpired はリフレッシュトークンが期限切れの場合に返されます。
var ErrTokenExpired = errors.New("refresh token expired")

// ErrInvalidToken はトークンが無効または存在しない場合に返されます。
var ErrInvalidToken = errors.New("invalid refresh token")

// UseCase はリフレッシュトークンの操作を提供します。
type UseCase interface {
	Create(ctx context.Context, in CreateInput) (CreateOutput, error)
	Refresh(ctx context.Context, in RefreshInput) (RefreshOutput, error)
	Revoke(ctx context.Context, in RevokeInput) error
}

type interactor struct {
	repo domaintoken.Repository
}

// NewUseCase は UseCase の実装を返します。
func NewUseCase(repo domaintoken.Repository) UseCase {
	return &interactor{repo: repo}
}

// Create はランダムなリフレッシュトークンを生成して DB に保存します。
// クライアントには平文トークンを返し、DB にはそのハッシュを保存します。
func (i *interactor) Create(ctx context.Context, in CreateInput) (CreateOutput, error) {
	plainToken, err := generateToken()
	if err != nil {
		return CreateOutput{}, fmt.Errorf("generate refresh token: %w", err)
	}

	expiresAt := time.Now().Add(refreshTokenTTL)
	if err := i.repo.Save(ctx, in.UserID, hashToken(plainToken), expiresAt); err != nil {
		return CreateOutput{}, fmt.Errorf("save refresh token: %w", err)
	}

	return CreateOutput{Token: plainToken, ExpiresAt: expiresAt}, nil
}

// Refresh はリフレッシュトークンを検証し、新しいアクセストークン発行に必要な UserID を返します。
func (i *interactor) Refresh(ctx context.Context, in RefreshInput) (RefreshOutput, error) {
	rt, err := i.repo.FindByToken(ctx, hashToken(in.Token))
	if err != nil {
		if errors.Is(err, domaintoken.ErrTokenNotFound) {
			return RefreshOutput{}, ErrInvalidToken
		}
		return RefreshOutput{}, fmt.Errorf("find refresh token: %w", err)
	}

	if time.Now().After(rt.ExpiresAt) {
		_ = i.repo.Delete(ctx, hashToken(in.Token))
		return RefreshOutput{}, ErrTokenExpired
	}

	return RefreshOutput{UserID: rt.UserID}, nil
}

// Revoke はリフレッシュトークンを DB から削除して無効化します。
func (i *interactor) Revoke(ctx context.Context, in RevokeInput) error {
	if err := i.repo.Delete(ctx, hashToken(in.Token)); err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// hashToken は平文トークンを SHA-256 でハッシュして返します。
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
