package receiptslist

import (
	"context"
	"errors"

	"receiptScan-backend/internal/domain/receipt"
)

// ErrInvalidUserID is returned when the user ID is invalid.
var ErrInvalidUserID = errors.New("userId must be positive")

// UseCase fetches receipts associated with a user ID.
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
	if in.UserID <= 0 {
		return Output{}, ErrInvalidUserID
	}

	receipts, err := i.repository.GetReceiptsByUserID(ctx, in.UserID)
	if err != nil {
		return Output{}, err
	}

	return Output{Receipts: receipts}, nil
}
