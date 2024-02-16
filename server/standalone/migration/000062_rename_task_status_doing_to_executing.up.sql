SELECT execute($$
    -- As we use the status as foreign key, we need to:
    --        1.  add the new name of the renamed ones + the new ones
    --        2. change compute tasks to use the new names of the renamed ones
    --        3. delete the previous values of the renamed ones
    INSERT INTO compute_task_statuses(status)
    VALUES ('STATUS_EXECUTING');

    UPDATE compute_tasks
    SET status = 'STATUS_EXECUTING'
    WHERE status = 'STATUS_DOING';

    DELETE FROM compute_task_statuses
    WHERE status  in ('STATUS_DOING');

    UPDATE events e
    SET asset = jsonb_set(asset, '{status}', to_jsonb('STATUS_EXECUTING'::text))
    WHERE asset_kind = 'ASSET_COMPUTE_TASK' AND asset->>'status' = 'STATUS_DOING';
    
$$) WHERE exists(SELECT 1 FROM compute_task_statuses WHERE status = 'STATUS_DOING');