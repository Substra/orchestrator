CREATE TABLE events (
    id UUID PRIMARY KEY,
    asset_key varchar(100) NOT NULL,
    channel varchar(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    event JSONB NOT NULL
);

CREATE INDEX ix_event_asset ON events(asset_key);
