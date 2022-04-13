CREATE OR REPLACE VIEW expanded_compute_tasks AS
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
