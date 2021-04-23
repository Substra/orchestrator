CREATE TABLE models (
    id UUID PRIMARY KEY,
    channel varchar(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    asset JSONB NOT NULL
);

CREATE INDEX ix_model_task ON models USING HASH ((asset->>'computeTaskKey'));
