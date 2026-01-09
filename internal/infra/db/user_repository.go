package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"receiptScan-backend/internal/domain/user"
)

// UserRepository handles persistence for users.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository initializes a repository for users.
func NewUserRepository(ctx context.Context, dsn string) (*UserRepository, error) {
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

	return &UserRepository{pool: pool}, nil
}

// Close closes the underlying connection pool.
func (r *UserRepository) Close() {
	r.pool.Close()
}

// Authenticate checks whether the provided credentials match.
func (r *UserRepository) Authenticate(ctx context.Context, userName string, passwordHash string) (bool, error) {
	var storedHash string
	if err := r.pool.QueryRow(ctx, `
SELECT password_hash
FROM users
WHERE user_name = $1
`, userName).Scan(&storedHash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("query user: %w", err)
	}

	return storedHash == passwordHash, nil
}

var _ user.Repository = (*UserRepository)(nil)

// Create inserts a user with hashed credentials.
func (r *UserRepository) Create(ctx context.Context, userName string, email, passwordHash string) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO users (user_name, email, password_hash)
VALUES ($1, $2, $3)
`, userName, email, passwordHash)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}
