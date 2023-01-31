ALTER TABLE algos
RENAME COLUMN algorithm TO functionAdress;

ALTER TABLE algos
RENAME TO functions;

ALTER TABLE compute_tasks
RENAME COLUMN algo_key TO function_key;

ALTER TABLE algo_outputs
RENAME COLUMN algo_key TO function_key;

ALTER TABLE algo_outputs
RENAME TO function_outputs;

ALTER TABLE algo_inputs
RENAME COLUMN algo_key TO function_key;

ALTER TABLE algo_inputs
RENAME TO function_inputs;

DROP VIEW IF EXISTS expanded_compute_tasks;

DROP VIEW IF EXISTS expanded_algos;
CREATE VIEW expanded_functions AS
SELECT 	key,
        name,
        description             AS description_address,
        desc_add.checksum       AS description_checksum,
        functionAdress          AS function_address,
        function_add.checksum   AS function_checksum,
	    permissions,
        owner,
        creation_date,
        metadata,
        channel
FROM functions
JOIN addressables desc_add ON functions.description = desc_add.storage_address
JOIN addressables function_add ON functions.functionAdress = function_add.storage_address;

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
FROM compute_tasks t
        LEFT JOIN (
    SELECT child_task_key, JSONB_AGG(parent_task_key) AS parent_task_keys
    FROM compute_task_parents
    GROUP BY child_task_key
) p ON p.child_task_key = t.key;


