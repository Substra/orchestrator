CREATE TABLE datasamples (
    id UUID PRIMARY KEY,
    channel varchar(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    asset JSONB NOT NULL
);
