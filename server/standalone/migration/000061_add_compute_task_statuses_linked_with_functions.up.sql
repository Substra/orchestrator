SELECT execute($$
    -- As we use the status as foreign key, we need to:
    --        1.  add the new name of the renamed ones
    --        2. change compute tasks to use the new names of the renamed ones
    --        3. delete the previous values of the renamed ones
    INSERT INTO compute_task_statuses(status)
    VALUES ('STATUS_WAITING_FOR_PARENT_TASKS'),
           ('STATUS_WAITING_FOR_EXECUTOR_SLOT');

    UPDATE compute_tasks
    SET status = 'STATUS_WAITING_FOR_PARENT_TASKS'
    WHERE status = 'STATUS_WAITING';

    UPDATE compute_tasks
    SET status = 'STATUS_WAITING_FOR_EXECUTOR_SLOT'
    WHERE status = 'STATUS_TODO';

    DELETE FROM compute_task_statuses
    WHERE status  in ('STATUS_TODO', 'STATUS_WAITING');
    
$$) WHERE NOT exists(SELECT 1 FROM compute_task_statuses WHERE status = 'STATUS_WAITING_FOR_PARENT_TASKS');