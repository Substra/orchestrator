SELECT execute($$
    ALTER TABLE performances
    ADD COLUMN asset jsonb;

    UPDATE performances
    SET asset = JSONB_BUILD_OBJECT(
        'computeTaskKey', compute_task_key,
        'metricKey', algo_key,
        'performanceValue', performance_value,
        'creationDate', to_rfc_3339(creation_date)
    );

    ALTER TABLE performances
    ALTER COLUMN asset SET NOT NULL;

    ALTER TABLE performances
    RENAME COLUMN compute_task_key TO compute_task_id;

    ALTER TABLE performances
    RENAME COLUMN algo_key TO metric_id;

    ALTER TABLE performances
    RENAME CONSTRAINT performances_compute_task_key_fkey TO performances_compute_task_id_fkey;

    ALTER TABLE performances
    RENAME CONSTRAINT performances_algo_key_fkey TO performances_metric_id_fkey;

    ALTER TABLE performances
    DROP COLUMN performance_value,
    DROP COLUMN creation_date;
$$) WHERE column_exists('public', 'performances', 'creation_date');
