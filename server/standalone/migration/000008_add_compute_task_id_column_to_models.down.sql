SELECT execute($$

    ALTER TABLE models DROP compute_task_id;
    CREATE INDEX ix_models_task ON models USING HASH ((asset->>'computeTaskKey'));

$$) WHERE NOT index_exists('ix_models_task');
