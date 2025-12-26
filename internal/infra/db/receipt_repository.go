package db

import (
	"context"
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
func (r *ReceiptRepository) Save(ctx context.Context, f receipt.FormattedReceipt) (err error) {
	purchaseDate, err := parsePurchaseDate(f.Date)
	if err != nil {
		return fmt.Errorf("parse date: %w", err)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var receiptID int64
	err = tx.QueryRow(ctx, `
INSERT INTO receipts (user_id, store, date, total)
VALUES ($1, $2, $3, $4)
RETURNING id
`, f.UserID, f.Store, purchaseDate, f.Total).Scan(&receiptID)
	if err != nil {
		return fmt.Errorf("insert receipt: %w", err)
	}

	for i, item := range f.Items {
		if _, err = tx.Exec(ctx, `
INSERT INTO items (receipt_id, name, price)
VALUES ($1, $2, $3)
`, receiptID, item.Name, item.Price); err != nil {
			return fmt.Errorf("insert item %d: %w", i, err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// GetItemsByReceiptID returns items linked to the given receipt ID.
func (r *ReceiptRepository) GetItemsByReceiptID(ctx context.Context, receiptID int64) ([]receipt.Item, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id, receipt_id, name, price
FROM items
WHERE receipt_id = $1
ORDER BY id
`, receiptID)
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	defer rows.Close()

	var items []receipt.Item
	for rows.Next() {
		var item receipt.Item
		if err := rows.Scan(&item.ID, &item.ReceiptID, &item.Name, &item.Price); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate items: %w", err)
	}

	return items, nil
}

// GetReceiptsByUserID returns receipts linked to the given user ID.
func (r *ReceiptRepository) GetReceiptsByUserID(ctx context.Context, userID int64) ([]receipt.Receipt, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id, user_id, store, date, total
FROM receipts
WHERE user_id = $1
ORDER BY date DESC, id DESC
`, userID)
	if err != nil {
		return nil, fmt.Errorf("query receipts: %w", err)
	}
	defer rows.Close()

	var receipts []receipt.Receipt
	for rows.Next() {
		var entry receipt.Receipt
		var purchaseDate time.Time
		if err := rows.Scan(&entry.ID, &entry.UserID, &entry.Store, &purchaseDate, &entry.Total); err != nil {
			return nil, fmt.Errorf("scan receipt: %w", err)
		}
		entry.Date = purchaseDate.Format(time.DateOnly)
		receipts = append(receipts, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate receipts: %w", err)
	}

	return receipts, nil
}

func parsePurchaseDate(dateStr string) (time.Time, error) {
	layouts := []string{
		time.DateOnly,
		"2006-01-02T15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}

	var lastErr error
	for _, layout := range layouts {
		t, err := time.Parse(layout, dateStr)
		if err == nil {
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
		}
		lastErr = err
	}

	return time.Time{}, lastErr
}
