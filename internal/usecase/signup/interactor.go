package signup

import (
	"context"
	"errors"
	"strings"

	"receiptScan-backend/internal/domain/email"
	"receiptScan-backend/internal/domain/user"
)

// ErrInvalidUserID is returned when the user ID is invalid.
var ErrInvalidUserID = errors.New("userId must be positive")

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
	mailer     email.Sender
}

// NewUseCase initializes a UseCase implementation.
func NewUseCase(repository user.Repository, mailer email.Sender) UseCase {
	return &interactor{repository: repository, mailer: mailer}
}

func (i *interactor) Execute(ctx context.Context, in Input) (Output, error) {
	if in.UserID <= 0 {
		return Output{}, ErrInvalidUserID
	}
	if in.Password == "" {
		return Output{}, ErrInvalidPassword
	}
	if strings.TrimSpace(in.Email) == "" {
		return Output{}, ErrInvalidEmail
	}

	hashed := user.HashPassword(in.UserID, in.Password)
	if err := i.repository.Create(ctx, in.UserID, in.Email, hashed); err != nil {
		return Output{}, err
	}

	if err := i.mailer.SendConfirmation(ctx, in.Email, in.UserID); err != nil {
		return Output{}, err
	}

	return Output{}, nil
}
