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
CREATE TABLE datasamples (
    id UUID PRIMARY KEY,
    channel varchar(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    asset JSONB NOT NULL
);
CREATE TABLE algos (
    id UUID PRIMARY KEY,
    channel varchar(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    asset JSONB NOT NULL
);

CREATE INDEX ix_algos_category ON algos USING HASH ((asset->>'category'));
CREATE TABLE datamanagers (
       id UUID PRIMARY KEY,
       channel varchar(100) NOT NULL,
       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
       asset JSONB NOT NULL
);
CREATE TABLE compute_tasks (
    id UUID PRIMARY KEY,
    channel varchar(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    asset JSONB NOT NULL
);

CREATE INDEX ix_compute_tasks_parents ON compute_tasks USING GIN ((asset->'parentTaskKeys'));
CREATE INDEX ix_compute_tasks_status ON compute_tasks USING HASH ((asset->>'status'));
CREATE INDEX ix_compute_tasks_category ON compute_tasks USING HASH ((asset->>'category'));
CREATE INDEX ix_compute_tasks_compute_plan_key ON compute_tasks USING HASH ((asset->>'computePlanKey'));
CREATE INDEX ix_compute_tasks_worker ON compute_tasks USING HASH ((asset->>'worker'));
CREATE TABLE models (
    id UUID PRIMARY KEY,
    channel varchar(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    asset JSONB NOT NULL
);

CREATE INDEX ix_model_task ON models USING HASH ((asset->>'computeTaskKey'));
CREATE INDEX ix_model_category ON models USING HASH ((asset->>'category'));
CREATE TABLE compute_plans (
    id UUID PRIMARY KEY,
    channel varchar(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    asset JSONB NOT NULL
);
CREATE TABLE performances (
    id UUID PRIMARY KEY,
    channel varchar(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    asset JSONB NOT NULL
);
CREATE TABLE events (
    id UUID PRIMARY KEY,
    asset_key varchar(100) NOT NULL,
    channel varchar(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    event JSONB NOT NULL
);

CREATE INDEX ix_event_asset ON events(asset_key);


