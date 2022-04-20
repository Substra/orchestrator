SELECT execute($$
    ALTER TABLE performances
    ADD COLUMN performance_value decimal,
    ADD COLUMN creation_date timestamptz;

    UPDATE performances
    SET performance_value = (asset ->> 'performanceValue')::decimal,
        creation_date = (asset ->> 'creationDate')::timestamptz;

    ALTER TABLE performances
    ALTER COLUMN performance_value SET NOT NULL,
    ALTER COLUMN creation_date SET NOT NULL;

    ALTER TABLE performances
    RENAME COLUMN compute_task_id TO compute_task_key;

    ALTER TABLE performances
    RENAME COLUMN metric_id TO algo_key;

    ALTER TABLE performances
    RENAME CONSTRAINT performances_compute_task_id_fkey TO performances_compute_task_key_fkey;

    ALTER TABLE performances
    RENAME CONSTRAINT performances_metric_id_fkey TO performances_algo_key_fkey;

    ALTER TABLE performances
    DROP COLUMN asset;
$$) WHERE column_exists('public', 'performances', 'asset');
