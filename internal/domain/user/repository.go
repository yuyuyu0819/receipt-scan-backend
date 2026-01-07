package user

import "context"

// Repository is the interface for user authentication.
type Repository interface {
	Authenticate(ctx context.Context, userID string, passwordHash string) (bool, error)
	Create(ctx context.Context, userID string, email, passwordHash string) error
}
