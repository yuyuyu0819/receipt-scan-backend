package items

import (
	"context"
	"errors"

	"receiptScan-backend/internal/domain/receipt"
)

// ErrInvalidReceiptID is returned when the receipt ID is zero or negative.
var ErrInvalidReceiptID = errors.New("receiptId must be positive")

// UseCase fetches items associated with a receipt ID.
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
	if in.ReceiptID <= 0 {
		return Output{}, ErrInvalidReceiptID
	}

	items, err := i.repository.GetItemsByReceiptID(ctx, in.ReceiptID)
	if err != nil {
		return Output{}, err
	}

	return Output{Items: items}, nil
}
