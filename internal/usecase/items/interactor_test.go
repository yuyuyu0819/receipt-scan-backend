package items

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"receiptScan-backend/internal/domain/receipt"
)

type stubRepository struct {
	items []receipt.Item
	err   error
}

func (s *stubRepository) Save(ctx context.Context, r receipt.FormattedReceipt) error {
	return nil
}

func (s *stubRepository) GetItemsByReceiptID(ctx context.Context, receiptID int64) ([]receipt.Item, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.items, nil
}

func (s *stubRepository) GetReceiptsByUserID(ctx context.Context, userID string) ([]receipt.Receipt, error) {
	return nil, nil
}

func TestExecute_InvalidID(t *testing.T) {
	uc := NewUseCase(&stubRepository{})

	_, err := uc.Execute(context.Background(), Input{ReceiptID: 0})
	if !errors.Is(err, ErrInvalidReceiptID) {
		t.Fatalf("expected ErrInvalidReceiptID, got %v", err)
	}
}

func TestExecute_ReturnsItems(t *testing.T) {
	expected := []receipt.Item{{ID: 1, ReceiptID: 10, Name: "apple", Price: 100}}
	uc := NewUseCase(&stubRepository{items: expected})

	out, err := uc.Execute(context.Background(), Input{ReceiptID: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(out.Items, expected) {
		t.Fatalf("expected %v, got %v", expected, out.Items)
	}
}

func TestExecute_RepositoryError(t *testing.T) {
	repoErr := errors.New("db error")
	uc := NewUseCase(&stubRepository{err: repoErr})

	_, err := uc.Execute(context.Background(), Input{ReceiptID: 1})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected %v, got %v", repoErr, err)
	}
}
