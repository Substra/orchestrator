UPDATE datamanagers
SET logs_permission = JSONB_BUILD_OBJECT('authorizedIds', JSON_BUILD_ARRAY(owner))
WHERE logs_permission = '{}'::jsonb;

UPDATE compute_tasks
SET logs_permission = JSONB_BUILD_OBJECT('authorizedIds', JSON_BUILD_ARRAY(owner))
WHERE logs_permission = '{}'::jsonb;
