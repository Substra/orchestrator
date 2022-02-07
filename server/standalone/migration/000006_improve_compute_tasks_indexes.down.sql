SELECT execute($$

    ALTER TABLE compute_tasks
        DROP compute_plan_key, /* this also drop the indexes */
        DROP status,
        DROP category,
        DROP worker;

    CREATE INDEX ix_compute_tasks_compute_plan_key ON compute_tasks USING HASH ((asset->>'computePlanKey'));
    CREATE INDEX ix_compute_tasks_category ON compute_tasks USING HASH ((asset->>'category'));
    CREATE INDEX ix_compute_tasks_worker ON compute_tasks USING HASH ((asset->>'worker'));
    CREATE INDEX ix_compute_tasks_status ON compute_tasks USING HASH ((asset->>'status'));

    DROP TABLE compute_task_parents; /* this also drop the indexes */

    CREATE INDEX ix_compute_tasks_parents ON compute_tasks USING GIN ((asset->'parentTaskKeys'));

$$) WHERE NOT index_exists('ix_compute_tasks_parents');
