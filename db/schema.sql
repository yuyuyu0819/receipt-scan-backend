-- Users table used for authentication.
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    user_name TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
ALTER TABLE IF EXISTS users
    ADD COLUMN IF NOT EXISTS user_name TEXT;
ALTER TABLE IF EXISTS users
    ADD COLUMN IF NOT EXISTS id BIGSERIAL;
ALTER TABLE IF EXISTS users
    ADD COLUMN IF NOT EXISTS email TEXT;
ALTER TABLE IF EXISTS users
    ADD COLUMN IF NOT EXISTS password_hash TEXT;
ALTER TABLE IF EXISTS users
    DROP COLUMN IF EXISTS password;
ALTER TABLE IF EXISTS users
    ADD CONSTRAINT users_user_name_key UNIQUE (user_name);
ALTER TABLE IF EXISTS users
    DROP CONSTRAINT IF EXISTS users_pkey;
ALTER TABLE IF EXISTS users
    ADD PRIMARY KEY (id);

-- Receipts table used by the application.
--
-- NOTE: Older versions stored items as a JSON column on receipts.
-- To avoid NOT NULL violations when migrating existing databases,
-- we drop that column if it exists before ensuring the current schema.
ALTER TABLE IF EXISTS receipts
    DROP COLUMN IF EXISTS items;
ALTER TABLE IF EXISTS receipts
    ADD COLUMN IF NOT EXISTS user_id BIGINT REFERENCES users(id);
ALTER TABLE IF EXISTS receipts
    DROP COLUMN IF EXISTS user_name;

CREATE TABLE IF NOT EXISTS receipts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    store TEXT NOT NULL,
    date DATE NOT NULL,
    total INTEGER NOT NULL CHECK (total >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Recommended index for querying receipts by purchase date.
CREATE INDEX IF NOT EXISTS idx_receipts_date ON receipts (date);
CREATE INDEX IF NOT EXISTS idx_receipts_user_id ON receipts (user_id);

-- Items table linked to receipts.
CREATE TABLE IF NOT EXISTS items (
    id BIGSERIAL PRIMARY KEY,
    receipt_id BIGINT NOT NULL REFERENCES receipts(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    price INTEGER NOT NULL CHECK (price >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_items_receipt_id ON items (receipt_id);
