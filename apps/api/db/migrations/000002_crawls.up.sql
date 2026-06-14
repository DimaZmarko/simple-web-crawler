-- Single-tenant for now: no user_id column. Per-user ownership is deferred to
-- spec 006 (auth), when a user_id FK + ownership filtering will be added.
CREATE TABLE crawls (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    seed_url    TEXT NOT NULL,
    max_depth   INT  NOT NULL CHECK (max_depth >= 0),
    max_pages   INT  NOT NULL CHECK (max_pages >= 1),
    status      TEXT NOT NULL CHECK (status = 'queued'),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Supports keyset pagination ordering: newest first by (created_at, id).
CREATE INDEX crawls_created_at_id_idx ON crawls (created_at DESC, id DESC);
