

SELECT execute($$
    DROP VIEW IF EXISTS expanded_compute_tasks;
    CREATE VIEW expanded_compute_tasks AS
    SELECT t.key AS key,
        t.channel AS channel,
        t.compute_plan_key AS compute_plan_key,
        t.status AS status,
        t.worker AS worker,
        t.owner as owner,
        t.rank AS rank,
        t.creation_date AS creation_date,
        t.logs_permission AS logs_permission,
        t.task_data AS task_data,
        t.metadata AS metadata,
        t.algo_key AS algo_key,
        COALESCE(p.parent_task_keys, '[]'::jsonb) AS parent_task_keys
    FROM compute_tasks t
             LEFT JOIN (
        SELECT child_task_key, JSONB_AGG(parent_task_key) AS parent_task_keys
        FROM compute_task_parents
        GROUP BY child_task_key
    ) p ON p.child_task_key = t.key;

    ALTER TABLE compute_tasks DROP COLUMN category;

    UPDATE events SET asset = asset #- '{category}'
    WHERE asset_kind = 'ASSET_COMPUTE_TASK';
$$) WHERE column_exists('public', 'compute_tasks', 'category');
