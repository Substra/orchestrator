SELECT execute($$

    ALTER TABLE failure_reports
    DROP CONSTRAINT ck_failure_reports_error_type_logs_address;

    UPDATE failure_reports
    SET logs_address = null
    WHERE logs_address = 'https://db-backfill/EMPTY_LOG_FILE';

    DELETE FROM addressables
    WHERE storage_address = 'https://db-backfill/EMPTY_LOG_FILE';

    ALTER TABLE failure_reports ADD CONSTRAINT error_type_logs_address_check CHECK (
        (error_type = 'ERROR_TYPE_EXECUTION' AND logs_address IS NOT NULL) OR
        (error_type != 'ERROR_TYPE_EXECUTION' AND logs_address IS NULL)
    );

$$) WHERE constraint_exists('failure_reports', 'ck_failure_reports_error_type_logs_address');
