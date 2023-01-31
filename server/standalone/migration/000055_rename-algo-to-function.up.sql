ALTER TABLE algos
RENAME TO functions;

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
       t.algo_key AS function_key,
       COALESCE(p.parent_task_keys, '[]'::jsonb) AS parent_task_keys
FROM compute_tasks t
         LEFT JOIN expanded_algos a ON a.key = t.algo_key
         LEFT JOIN (
    SELECT child_task_key, JSONB_AGG(parent_task_key ORDER BY position) AS parent_task_keys
    FROM compute_task_parents
    GROUP BY child_task_key
) p ON p.child_task_key = t.key;


DROP VIEW IF EXISTS expanded_algos;
CREATE VIEW expanded_functions AS
SELECT 	key,
        name,
        description_address AS description_address,
        description_checksum AS description_checksum,
        algorithm_address AS function_address,
        algorithm_checksum AS function_checksum,
	permissions,
        owner,
        creation_date,
        metadata,
        channel
FROM functions;
