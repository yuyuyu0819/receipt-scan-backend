-- Receipts table used by the application.
--
-- NOTE: Older versions stored items as a JSON column on receipts.
-- To avoid NOT NULL violations when migrating existing databases,
-- we drop that column if it exists before ensuring the current schema.
ALTER TABLE IF EXISTS receipts
    DROP COLUMN IF EXISTS items;

CREATE TABLE IF NOT EXISTS receipts (
    id BIGSERIAL PRIMARY KEY,
    store TEXT NOT NULL,
    date DATE NOT NULL,
    total INTEGER NOT NULL CHECK (total >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Recommended index for querying receipts by purchase date.
CREATE INDEX IF NOT EXISTS idx_receipts_date ON receipts (date);

-- Items table linked to receipts.
CREATE TABLE IF NOT EXISTS items (
    id BIGSERIAL PRIMARY KEY,
    receipt_id BIGINT NOT NULL REFERENCES receipts(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    price INTEGER NOT NULL CHECK (price >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_items_receipt_id ON items (receipt_id);
