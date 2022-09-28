SELECT execute($$

    ALTER TABLE failure_reports
    DROP CONSTRAINT error_type_logs_address_check;

    INSERT INTO addressables(storage_address, checksum)
    VALUES ('https://db-backfill/EMPTY_LOG_FILE', '0000000000000000000000000000000000000000000000000000000000000000');

    UPDATE failure_reports
    SET logs_address = 'https://db-backfill/EMPTY_LOG_FILE'
    WHERE error_type = 'ERROR_TYPE_BUILD';

    ALTER TABLE failure_reports ADD CONSTRAINT ck_failure_reports_error_type_logs_address CHECK (
        (    (error_type = 'ERROR_TYPE_EXECUTION' OR error_type = 'ERROR_TYPE_BUILD') AND logs_address IS NOT NULL) OR
        (NOT (error_type = 'ERROR_TYPE_EXECUTION' OR error_type = 'ERROR_TYPE_BUILD') AND logs_address IS NULL)
    );

$$) WHERE NOT constraint_exists('failure_reports', 'ck_failure_reports_error_type_logs_address');
