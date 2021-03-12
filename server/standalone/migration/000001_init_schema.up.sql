CREATE TABLE nodes (
       id varchar(100),
       channel varchar(100),
       PRIMARY KEY (id, channel)
);
CREATE TABLE objectives (
       id UUID PRIMARY KEY,
       channel varchar(100) NOT NULL,
       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
       asset JSONB NOT NULL
);
