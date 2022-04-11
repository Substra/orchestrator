SELECT execute($$
    CREATE TABLE compute_task_categories (
        category varchar(50) PRIMARY KEY
    );

    INSERT INTO compute_task_categories(category)
    VALUES ('TASK_TRAIN'),
           ('TASK_AGGREGATE'),
           ('TASK_COMPOSITE'),
           ('TASK_TEST');

    CREATE TABLE compute_task_statuses (
        status varchar(100) PRIMARY KEY
    );

    INSERT INTO compute_task_statuses(status)
    VALUES ('STATUS_WAITING'),
           ('STATUS_TODO'),
           ('STATUS_DOING'),
           ('STATUS_DONE'),
           ('STATUS_CANCELED'),
           ('STATUS_FAILED');

    ALTER TABLE compute_tasks
    ADD COLUMN algo_key        uuid REFERENCES algos (key),
    ADD COLUMN owner           varchar(100),
    ADD COLUMN rank            integer,
    ADD COLUMN creation_date   timestamptz,
    ADD COLUMN logs_permission jsonb,
    ADD COLUMN task_data       jsonb,
    ADD COLUMN metadata        jsonb,
    ADD CONSTRAINT compute_tasks_category_fkey FOREIGN KEY (CATEGORY) REFERENCES compute_task_categories (CATEGORY),
    ADD CONSTRAINT compute_tasks_status_fkey FOREIGN KEY (status) REFERENCES compute_task_statuses (status),
    ADD CONSTRAINT compute_tasks_owner_channel_fkey FOREIGN KEY (owner, channel) REFERENCES nodes (id, channel),
    ADD CONSTRAINT compute_tasks_worker_channel_fkey FOREIGN KEY (worker, channel) REFERENCES nodes (id, channel);

    UPDATE compute_tasks
    SET algo_key        = (asset -> 'algo' ->> 'key')::uuid,
        owner           = asset ->> 'owner',
        rank            = COALESCE((asset ->> 'rank')::integer, 0),
        creation_date   = (asset ->> 'creationDate')::timestamptz,
        logs_permission = COALESCE(asset -> 'logsPermission', '{}'::jsonb),
        task_data       = CASE category
                              WHEN 'TASK_TRAIN' THEN JSONB_BUILD_OBJECT('train', asset -> 'train')
                              WHEN 'TASK_AGGREGATE' THEN JSONB_BUILD_OBJECT('aggregate', asset -> 'aggregate')
                              WHEN 'TASK_COMPOSITE' THEN JSONB_BUILD_OBJECT('composite', asset -> 'composite')
                              WHEN 'TASK_TEST' THEN JSONB_BUILD_OBJECT('test', asset -> 'test') END,
        metadata        = COALESCE(asset -> 'metadata', '{}'::jsonb);

    ALTER TABLE compute_tasks
    ALTER COLUMN algo_key SET NOT NULL,
    ALTER COLUMN owner SET NOT NULL,
    ALTER COLUMN rank SET NOT NULL,
    ALTER COLUMN creation_date SET NOT NULL,
    ALTER COLUMN logs_permission SET NOT NULL,
    ALTER COLUMN task_data SET NOT NULL,
    ALTER COLUMN metadata SET NOT NULL,
    ALTER COLUMN metadata SET DEFAULT '{}'::jsonb;

    ALTER TABLE compute_tasks
    RENAME COLUMN compute_plan_id TO compute_plan_key;

    ALTER INDEX ix_compute_tasks_compute_plan_id RENAME TO ix_compute_tasks_compute_plan_key;

    ALTER TABLE compute_tasks
    RENAME COLUMN id TO key;

    ALTER TABLE compute_task_parents
    RENAME COLUMN child_task_id TO child_task_key;

    ALTER TABLE compute_task_parents
    RENAME CONSTRAINT compute_task_parents_child_task_id_fkey TO compute_task_parents_child_task_key_fkey;

    ALTER TABLE compute_task_parents
    RENAME COLUMN parent_task_id TO parent_task_key;

    ALTER TABLE compute_task_parents
    RENAME CONSTRAINT compute_task_parents_parent_task_id_fkey TO compute_task_parents_parent_task_key_fkey;

    ALTER INDEX ix_compute_task_parents_child_task_id RENAME TO ix_compute_task_parents_child_task_key;
    ALTER INDEX ix_compute_task_parents_parent_task_id RENAME TO ix_compute_task_parents_parent_task_key;

    ALTER TABLE compute_tasks
    DROP COLUMN asset;

    CREATE VIEW expanded_compute_tasks AS
    SELECT t.key AS key,
           t.channel AS channel,
           t.compute_plan_key AS compute_plan_key,
           t.status AS status,
           t.category AS category,
           t.worker AS worker,
           t.owner as owner,
           t.rank AS rank,
           t.creation_date AS creation_date,
           t.logs_permission AS logs_permission,
           t.task_data AS task_data,
           t.metadata AS metadata,
           a.key AS algo_key,
           a.name AS algo_name,
           a.category AS algo_category,
           a.description_address AS algo_description_address,
           a.description_checksum AS algo_description_checksum,
           a.algorithm_address AS algo_algorithm_address,
           a.algorithm_checksum AS algo_algorithm_checksum,
           a.permissions AS algo_permissions,
           a.owner AS algo_owner,
           a.creation_date AS algo_creation_date,
           a.metadata AS algo_metadata,
           COALESCE(p.parent_task_keys, '[]'::jsonb) AS parent_task_keys
    FROM compute_tasks t
    LEFT JOIN expanded_algos a ON a.key = t.algo_key
    LEFT JOIN (
        SELECT child_task_key, JSONB_AGG(parent_task_key) AS parent_task_keys
        FROM compute_task_parents
        GROUP BY child_task_key
    ) p ON p.child_task_key = t.key;
$$) WHERE NOT view_exists('public', 'expanded_compute_tasks');
