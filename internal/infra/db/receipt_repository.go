package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"receiptScan-backend/internal/domain/receipt"
)

// ReceiptRepository は PostgreSQL への永続化を担います。
type ReceiptRepository struct {
	pool *pgxpool.Pool
}

// NewReceiptRepository は接続プールを初期化し、レポジトリを返します。
func NewReceiptRepository(ctx context.Context, dsn string) (*ReceiptRepository, error) {
	if dsn == "" {
		return nil, errors.New("dsn is empty")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &ReceiptRepository{pool: pool}, nil
}

// Close は内部の接続プールをクローズします。
func (r *ReceiptRepository) Close() {
	r.pool.Close()
}

// Save は整形済みレシートを receipts テーブルに保存します。
func (r *ReceiptRepository) Save(ctx context.Context, f receipt.FormattedReceipt) error {
	itemsJSON, err := json.Marshal(f.Items)
	if err != nil {
		return fmt.Errorf("marshal items: %w", err)
	}

	purchaseDate, err := time.Parse(time.DateOnly, f.Date)
	if err != nil {
		return fmt.Errorf("parse date: %w", err)
	}

	_, err = r.pool.Exec(ctx, `
INSERT INTO receipts (store, date, total, items)
VALUES ($1, $2, $3, $4)
`, f.Store, purchaseDate, f.Total, string(itemsJSON))
	if err != nil {
		return fmt.Errorf("insert receipt: %w", err)
	}

	return nil
}
