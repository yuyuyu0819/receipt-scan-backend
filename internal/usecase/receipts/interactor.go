package receipts

import (
	"context"

	"receiptScan-backend/internal/domain/receipt"
)

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
	if err := i.repository.Save(ctx, in.Receipt); err != nil {
		return Output{}, err
	}

	return Output{}, nil
}
