SELECT execute($$
    ALTER TABLE compute_tasks DROP COLUMN category;
    DROP TABLE compute_task_categories;

    UPDATE events SET asset = asset - 'category'
    WHERE asset_kind = 'ASSET_COMPUTE_TASK';
$$) WHERE column_exists('public', 'compute_tasks', 'category');
