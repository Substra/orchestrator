CREATE TABLE nodes (
       id varchar(100) PRIMARY KEY
);
CREATE TABLE objectives (
       id UUID PRIMARY KEY,
       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
       asset JSONB NOT NULL
);
