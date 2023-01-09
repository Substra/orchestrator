SELECT execute($$
    DROP VIEW IF EXISTS expanded_compute_tasks;
    UPDATE events
    SET asset = asset - 'parentTaskKeys'
    WHERE asset_kind = 'ASSET_COMPUTE_TASK';
$$) WHERE view_exists('public', 'expanded_compute_tasks');
