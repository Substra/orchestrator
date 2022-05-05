SELECT execute($$
    ALTER TABLE compute_tasks
    ADD COLUMN asset JSONB;

    UPDATE compute_tasks t
    SET asset = JSONB_BUILD_OBJECT(
            'key', t.key,
            'category', t.category,
            'algo', build_algo_jsonb(
                    e.algo_key,
                    e.algo_name,
                    e.algo_category,
                    e.algo_description_checksum,
                    e.algo_description_address,
                    e.algo_algorithm_checksum,
                    e.algo_algorithm_address,
                    e.algo_permissions,
                    e.algo_owner,
                    e.algo_creation_date,
                    e.algo_metadata
                ),
            'owner', t.owner,
            'computePlanKey', t.compute_plan_key,
            'parentTaskKeys', e.parent_task_keys,
            'rank', t.rank,
            'status', t.status,
            'worker', t.worker,
            'creationDate', to_rfc_3339(t.creation_date),
            'logsPermission', t.logs_permission,
            'metadata', t.metadata
        ) || t.task_data
    FROM expanded_compute_tasks e
    WHERE t.key = e.key;

    ALTER TABLE compute_tasks
    ALTER COLUMN asset SET NOT NULL;

    DROP VIEW expanded_compute_tasks;

    ALTER INDEX ix_compute_task_parents_child_task_key RENAME TO ix_compute_task_parents_child_task_id;
    ALTER INDEX ix_compute_task_parents_parent_task_key RENAME TO ix_compute_task_parents_parent_task_id;

    ALTER TABLE compute_task_parents
    RENAME COLUMN parent_task_key TO parent_task_id;

    ALTER TABLE compute_task_parents
    RENAME CONSTRAINT compute_task_parents_parent_task_key_fkey TO compute_task_parents_parent_task_id_fkey;

    ALTER TABLE compute_task_parents
    RENAME COLUMN child_task_key TO child_task_id;

    ALTER TABLE compute_task_parents
    RENAME CONSTRAINT compute_task_parents_child_task_key_fkey TO compute_task_parents_child_task_id_fkey;

    ALTER TABLE compute_tasks
    RENAME COLUMN key TO id;

    ALTER INDEX ix_compute_tasks_compute_plan_key RENAME TO ix_compute_tasks_compute_plan_id;

    ALTER TABLE compute_tasks
    RENAME COLUMN compute_plan_key TO compute_plan_id;

    ALTER TABLE compute_tasks
    DROP CONSTRAINT compute_tasks_category_fkey,
    DROP CONSTRAINT compute_tasks_status_fkey,
    DROP CONSTRAINT compute_tasks_owner_channel_fkey,
    DROP CONSTRAINT compute_tasks_worker_channel_fkey,
    DROP COLUMN algo_key,
    DROP COLUMN owner,
    DROP COLUMN rank,
    DROP COLUMN creation_date,
    DROP COLUMN logs_permission,
    DROP COLUMN task_data,
    DROP COLUMN metadata;

    DROP TABLE compute_task_statuses;
    DROP TABLE compute_task_categories;
$$) WHERE table_exists('public', 'compute_task_categories');
