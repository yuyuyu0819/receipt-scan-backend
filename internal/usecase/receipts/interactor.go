package receipts

import (
	"context"
	"errors"
	"strings"

	"receiptScan-backend/internal/domain/receipt"
)

// ErrInvalidUserName is returned when the user name is invalid.
var ErrInvalidUserName = errors.New("userName must not be empty")

// UseCase saves formatted receipts.
type UseCase interface {
	Execute(ctx context.Context, in Input) (Output, error)
}

type interactor struct {
	repository receipt.Repository
}

// NewUseCase initializes a UseCase implementation.
func NewUseCase(repository receipt.Repository) UseCase {
	return &interactor{repository: repository}
}

func (i *interactor) Execute(ctx context.Context, in Input) (Output, error) {
	if strings.TrimSpace(in.Receipt.UserName) == "" {
		return Output{}, ErrInvalidUserName
	}
	if err := i.repository.Save(ctx, in.Receipt); err != nil {
		return Output{}, err
	}

	return Output{}, nil
}
