package user

import "context"

// Repository is the interface for user authentication.
type Repository interface {
	Authenticate(ctx context.Context, userName string, password string) (int64, bool, error)
	Create(ctx context.Context, userName string, email, passwordHash string) (int64, error)
}
