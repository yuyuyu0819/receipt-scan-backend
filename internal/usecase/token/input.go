package token

// CreateInput は新規リフレッシュトークン発行の入力です。
type CreateInput struct {
	UserID int64
}

// RefreshInput はアクセストークン再発行の入力です。
type RefreshInput struct {
	Token string
}

// RevokeInput はリフレッシュトークン失効の入力です。
type RevokeInput struct {
	Token string
}
