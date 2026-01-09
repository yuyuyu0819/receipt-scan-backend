package login

import (
	"context"
	"errors"
	"strings"

	"receiptScan-backend/internal/domain/user"
)

// ErrInvalidUserName is returned when the user name is invalid.
var ErrInvalidUserName = errors.New("userName must not be empty")

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
	if strings.TrimSpace(in.UserName) == "" {
		return Output{}, ErrInvalidUserName
	}
	if in.Password == "" {
		return Output{}, ErrInvalidPassword
	}

	hashed := user.HashPassword(in.UserName, in.Password)
	userID, ok, err := i.repository.Authenticate(ctx, in.UserName, hashed)
	if err != nil {
		return Output{}, err
	}
	if !ok {
		return Output{}, ErrAuthenticationFailed
	}

	return Output{UserID: userID}, nil
}
