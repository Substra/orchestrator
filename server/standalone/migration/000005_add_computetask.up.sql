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
