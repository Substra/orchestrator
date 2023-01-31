ALTER TABLE algos
RENAME TO functions;

ALTER TABLE compute_tasks
RENAME COLUMN algo_key TO function_key;

DROP VIEW IF EXISTS expanded_compute_tasks;

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
       t.metadata AS metadata,
       t.function_key AS function_key,
       COALESCE(p.parent_task_keys, '[]'::jsonb) AS parent_task_keys
FROM compute_tasks t;



