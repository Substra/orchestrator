DROP VIEW IF EXISTS expanded_compute_tasks;
UPDATE events
SET asset = asset - 'parent_task_keys'
WHERE asset_kind = 'ASSET_COMPUTE_TASK';
