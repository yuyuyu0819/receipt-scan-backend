package login

import (
	"context"
	"errors"

	"receiptScan-backend/internal/domain/user"
)

// ErrInvalidUserID is returned when the user ID is invalid.
var ErrInvalidUserID = errors.New("userId must be positive")

// ErrInvalidPassword is returned when the password is empty.
var ErrInvalidPassword = errors.New("password must not be empty")

// ErrAuthenticationFailed is returned when credentials are invalid.
var ErrAuthenticationFailed = errors.New("invalid credentials")

// UseCase performs authentication.
type UseCase interface {
	Execute(ctx context.Context, in Input) (Output, error)
}

type interactor struct {
	repository user.Repository
}

// NewUseCase initializes a UseCase implementation.
func NewUseCase(repository user.Repository) UseCase {
	return &interactor{repository: repository}
}

func (i *interactor) Execute(ctx context.Context, in Input) (Output, error) {
	if in.UserID <= 0 {
		return Output{}, ErrInvalidUserID
	}
	if in.Password == "" {
		return Output{}, ErrInvalidPassword
	}

	hashed := user.HashPassword(in.UserID, in.Password)
	ok, err := i.repository.Authenticate(ctx, in.UserID, hashed)
	if err != nil {
		return Output{}, err
	}
	if !ok {
		return Output{}, ErrAuthenticationFailed
	}

	return Output{}, nil
}
