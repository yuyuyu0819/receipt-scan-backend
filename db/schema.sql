-- Receipts table used by the application.
CREATE TABLE IF NOT EXISTS receipts (
    id BIGSERIAL PRIMARY KEY,
    store TEXT NOT NULL,
    date DATE NOT NULL,
    total INTEGER NOT NULL CHECK (total >= 0),
    items JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Recommended index for querying receipts by purchase date.
CREATE INDEX IF NOT EXISTS idx_receipts_date ON receipts (date);
