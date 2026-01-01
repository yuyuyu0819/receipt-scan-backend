package signup

import (
	"context"
	"errors"
	"strings"

	"receiptScan-backend/internal/domain/recaptcha"
	"receiptScan-backend/internal/domain/user"
)

// ErrInvalidUserID is returned when the user ID is invalid.
var ErrInvalidUserID = errors.New("userId must not be empty")

// ErrInvalidPassword is returned when the password is empty.
var ErrInvalidPassword = errors.New("password must not be empty")

// ErrInvalidEmail is returned when the email is invalid.
var ErrInvalidEmail = errors.New("email must not be empty")

// ErrInvalidRecaptchaToken is returned when the token is empty.
var ErrInvalidRecaptchaToken = errors.New("recaptchaToken must not be empty")

// ErrRecaptchaFailed is returned when verification fails.
var ErrRecaptchaFailed = errors.New("recaptcha verification failed")

// UseCase performs signup.
type UseCase interface {
	Execute(ctx context.Context, in Input) (Output, error)
}

type interactor struct {
	repository user.Repository
	verifier   recaptcha.Verifier
}

// NewUseCase initializes a UseCase implementation.
func NewUseCase(repository user.Repository, verifier recaptcha.Verifier) UseCase {
	return &interactor{repository: repository, verifier: verifier}
}

func (i *interactor) Execute(ctx context.Context, in Input) (Output, error) {
	if strings.TrimSpace(in.UserID) == "" {
		return Output{}, ErrInvalidUserID
	}
	if in.Password == "" {
		return Output{}, ErrInvalidPassword
	}
	if strings.TrimSpace(in.Email) == "" {
		return Output{}, ErrInvalidEmail
	}
	if strings.TrimSpace(in.RecaptchaToken) == "" {
		return Output{}, ErrInvalidRecaptchaToken
	}

	if err := i.verifier.Verify(ctx, in.RecaptchaToken); err != nil {
		return Output{}, ErrRecaptchaFailed
	}

	hashed := user.HashPassword(in.UserID, in.Password)
	if err := i.repository.Create(ctx, in.UserID, in.Email, hashed); err != nil {
		return Output{}, err
	}

	return Output{}, nil
}
