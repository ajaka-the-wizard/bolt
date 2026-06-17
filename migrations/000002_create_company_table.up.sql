CREATE TABLE IF NOT EXISTS company(
    name             TEXT NOT NULL,
    address1         TEXT NOT NULL,
    address2         TEXT,
    city             TEXT NOT NULL,
    state            TEXT NOT NULL,
    postal_code      TEXT NOT NULL,
    country          TEXT NOT NULL,
    phone            TEXT NOT NULL,
    email            TEXT NOT NULL UNIQUE,
    website          TEXT NOT NULL,
    tax_id           TEXT NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
