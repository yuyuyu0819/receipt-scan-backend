package user

import "context"

// Repository is the interface for user authentication.
type Repository interface {
	Authenticate(ctx context.Context, userID int64, passwordHash string) (bool, error)
	Create(ctx context.Context, userID int64, email, passwordHash string) error
}
