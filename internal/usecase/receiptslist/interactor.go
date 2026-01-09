package receiptslist

import (
	"context"
	"errors"
	"strings"

	"receiptScan-backend/internal/domain/receipt"
)

// ErrInvalidUserName is returned when the user name is invalid.
var ErrInvalidUserName = errors.New("userName must not be empty")

// UseCase fetches receipts associated with a user name.
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
	if strings.TrimSpace(in.UserName) == "" {
		return Output{}, ErrInvalidUserName
	}

	receipts, err := i.repository.GetReceiptsByUserID(ctx, in.UserName)
	if err != nil {
		return Output{}, err
	}

	return Output{Receipts: receipts}, nil
}
