SELECT execute($$ /* This makes the migration idempotent (see WHERE condition at the bottom of this file) */

    /* Move compute_plan_key, status, category, and worker to their own columns */

    ALTER TABLE compute_tasks
        ADD compute_plan_key UUID NULL,
        ADD status varchar(100) NULL,
        ADD category varchar(100) NULL,
        ADD worker varchar(100) NULL;

    UPDATE compute_tasks SET
        compute_plan_key = (asset->>'computePlanKey')::uuid,
        status = (asset->>'status'),
        category = (asset->>'category'),
        worker = (asset->>'worker');

    ALTER TABLE compute_tasks ALTER COLUMN compute_plan_key SET NOT NULL;
    ALTER TABLE compute_tasks ALTER COLUMN status SET NOT NULL;
    ALTER TABLE compute_tasks ALTER COLUMN category SET NOT NULL;
    ALTER TABLE compute_tasks ALTER COLUMN worker SET NOT NULL;

    DROP INDEX ix_compute_tasks_compute_plan_key;
    DROP INDEX ix_compute_tasks_category;
    DROP INDEX ix_compute_tasks_worker;
    DROP INDEX ix_compute_tasks_status;

    CREATE INDEX ix_compute_tasks_compute_plan_key ON compute_tasks (compute_plan_key);
    CREATE INDEX ix_compute_tasks_category ON compute_tasks (category);
    CREATE INDEX ix_compute_tasks_worker ON compute_tasks (worker);
    CREATE INDEX ix_compute_tasks_status ON compute_tasks (status);

    /* Move parentTaskKeys to its own table */

    CREATE TABLE compute_task_parents (
        child_task_id UUID REFERENCES compute_tasks (id),
        parent_task_id UUID REFERENCES compute_tasks (id),
        position integer NOT NULL,
        PRIMARY KEY(child_task_id, parent_task_id, position)
    );

    INSERT INTO compute_task_parents(child_task_id, parent_task_id, position)
    SELECT id, parent_task_key::uuid, ROW_NUMBER() OVER(PARTITION BY id) AS position
    FROM compute_tasks, jsonb_array_elements_text(asset->'parentTaskKeys') AS parent_task_key;

    CREATE INDEX ix_compute_task_parents_child_task_id ON compute_task_parents (child_task_id);
    CREATE INDEX ix_compute_task_parents_parent_task_id ON compute_task_parents (parent_task_id);

    DROP INDEX ix_compute_tasks_parents;

$$) WHERE NOT column_exists('public', 'compute_tasks', 'compute_plan_key');
