CREATE TABLE nodes (
       id varchar(100),
       channel varchar(100),
       asset JSONB NOT NULL,
       PRIMARY KEY (id, channel)
);
CREATE INDEX ix_nodes_creation ON nodes ((asset->>'creationDate'));

CREATE TABLE objectives (
       id UUID PRIMARY KEY,
       channel varchar(100) NOT NULL,
       asset JSONB NOT NULL
);
CREATE INDEX ix_objectives_creation ON objectives ((asset->>'creationDate'));

CREATE TABLE datasamples (
    id UUID PRIMARY KEY,
    channel varchar(100) NOT NULL,
    asset JSONB NOT NULL
);
CREATE INDEX ix_datasamples_creation ON datasamples ((asset->>'creationDate'));

CREATE TABLE algos (
    id UUID PRIMARY KEY,
    channel varchar(100) NOT NULL,
    asset JSONB NOT NULL
);
CREATE INDEX ix_algos_category ON algos USING HASH ((asset->>'category'));
CREATE INDEX ix_algos_creation ON algos ((asset->>'creationDate'));

CREATE TABLE datamanagers (
       id UUID PRIMARY KEY,
       channel varchar(100) NOT NULL,
       asset JSONB NOT NULL
);
CREATE INDEX ix_datamanagers_creation ON datamanagers ((asset->>'creationDate'));

CREATE TABLE compute_tasks (
    id UUID PRIMARY KEY,
    channel varchar(100) NOT NULL,
    asset JSONB NOT NULL
);
CREATE INDEX ix_compute_tasks_parents ON compute_tasks USING GIN ((asset->'parentTaskKeys'));
CREATE INDEX ix_compute_tasks_status ON compute_tasks USING HASH ((asset->>'status'));
CREATE INDEX ix_compute_tasks_category ON compute_tasks USING HASH ((asset->>'category'));
CREATE INDEX ix_compute_tasks_compute_plan_key ON compute_tasks USING HASH ((asset->>'computePlanKey'));
CREATE INDEX ix_compute_tasks_worker ON compute_tasks USING HASH ((asset->>'worker'));
CREATE INDEX ix_compute_tasks_test_objective_key ON compute_tasks USING HASH ((asset->'test'->>'objectiveKey'));
CREATE INDEX ix_compute_tasks_creation ON compute_tasks ((asset->>'creationDate'));

CREATE TABLE models (
    id UUID PRIMARY KEY,
    channel varchar(100) NOT NULL,
    asset JSONB NOT NULL
);
CREATE INDEX ix_models_task ON models USING HASH ((asset->>'computeTaskKey'));
CREATE INDEX ix_models_category ON models USING HASH ((asset->>'category'));
CREATE INDEX ix_models_creation ON models ((asset->>'creationDate'));

CREATE TABLE compute_plans (
    id UUID PRIMARY KEY,
    channel varchar(100) NOT NULL,
    asset JSONB NOT NULL
);
CREATE INDEX ix_compute_plans_creation ON compute_plans ((asset->>'creationDate'));

CREATE TABLE performances (
    id UUID PRIMARY KEY,
    channel varchar(100) NOT NULL,
    asset JSONB NOT NULL
);
CREATE INDEX ix_performances_compute_task_key ON performances USING HASH ((asset->>'computeTaskKey'));
CREATE INDEX ix_performances_creation ON performances ((asset->>'creationDate'));

CREATE TABLE events (
    id UUID PRIMARY KEY,
    asset_key varchar(100) NOT NULL,
    channel varchar(100) NOT NULL,
    event JSONB NOT NULL
);
CREATE INDEX ix_event_asset ON events(asset_key);
CREATE INDEX ix_event_creation ON events((event->>'timestamp'));
