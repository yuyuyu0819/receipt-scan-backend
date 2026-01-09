package signup

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

// ErrInvalidEmail is returned when the email is invalid.
var ErrInvalidEmail = errors.New("email must not be empty")

// UseCase performs signup.
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
	if strings.TrimSpace(in.Email) == "" {
		return Output{}, ErrInvalidEmail
	}
	hashed := user.HashPassword(in.UserName, in.Password)
	if err := i.repository.Create(ctx, in.UserName, in.Email, hashed); err != nil {
		return Output{}, err
	}

	return Output{}, nil
}
