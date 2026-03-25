package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

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
func (r *UserRepository) Authenticate(ctx context.Context, userName string, password string) (int64, bool, error) {
	var storedHash string
	var userID int64
	if err := r.pool.QueryRow(ctx, `
SELECT id, password_hash
FROM users
WHERE user_name = $1
`, userName).Scan(&userID, &storedHash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("query user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err != nil {
		return 0, false, nil
	}
	return userID, true, nil
}

var _ user.Repository = (*UserRepository)(nil)

// Create inserts a user with hashed credentials.
func (r *UserRepository) Create(ctx context.Context, userName string, email, passwordHash string) (int64, error) {
	var userID int64
	if err := r.pool.QueryRow(ctx, `
INSERT INTO users (user_name, email, password_hash)
VALUES ($1, $2, $3)
RETURNING id
`, userName, email, passwordHash).Scan(&userID); err != nil {
		return 0, fmt.Errorf("insert user: %w", err)
	}

	return userID, nil
}
