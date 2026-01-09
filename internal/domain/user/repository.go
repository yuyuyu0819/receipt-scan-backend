package user

import "context"

// Repository is the interface for user authentication.
type Repository interface {
	Authenticate(ctx context.Context, userName string, passwordHash string) (bool, error)
	Create(ctx context.Context, userName string, email, passwordHash string) error
}
