SELECT execute($$

    DROP INDEX ix_models_task;

    ALTER TABLE models ADD compute_task_id UUID NULL REFERENCES compute_tasks (id);
    UPDATE models SET compute_task_id = (asset->>'computeTaskKey')::uuid;
    ALTER TABLE models ALTER COLUMN compute_task_id SET NOT NULL;

    CREATE INDEX ix_models_compute_task_id ON models USING HASH (compute_task_id);

$$) WHERE NOT column_exists('public', 'models', 'compute_task_id');
