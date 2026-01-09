package receipts

import (
	"context"
	"errors"

	"receiptScan-backend/internal/domain/receipt"
)

// ErrInvalidUserID is returned when the user ID is invalid.
var ErrInvalidUserID = errors.New("userId must be positive")

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
	if in.Receipt.UserID <= 0 {
		return Output{}, ErrInvalidUserID
	}
	if err := i.repository.Save(ctx, in.Receipt); err != nil {
		return Output{}, err
	}

	return Output{}, nil
}
