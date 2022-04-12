SELECT execute($$
    DROP VIEW expanded_failure_reports;

    ALTER TABLE failure_reports
    ADD COLUMN asset JSONB;

    UPDATE failure_reports r
    SET asset = JSONB_BUILD_OBJECT(
            'computeTaskKey', r.compute_task_key,
            'errorType', r.error_type,
            'logsAddress', build_addressable_jsonb(a.checksum, a.storage_address),
            'creationDate', to_rfc_3339(r.creation_date),
            'owner', r.owner
        )
    FROM addressables a
    WHERE r.logs_address = a.storage_address;

    ALTER TABLE failure_reports
    ALTER COLUMN asset SET NOT NULL;

    ALTER TABLE failure_reports
    RENAME COLUMN compute_task_key TO compute_task_id;

    ALTER TABLE failure_reports
    RENAME CONSTRAINT failure_reports_compute_task_key_fkey TO failure_reports_compute_task_id_fkey;

    ALTER TABLE failure_reports
    DROP CONSTRAINT failure_reports_logs_address_fkey;

    DELETE FROM addressables
    WHERE storage_address IN (SELECT logs_address FROM failure_reports);

    ALTER TABLE failure_reports
    DROP CONSTRAINT failure_reports_owner_channel_fkey,
    DROP CONSTRAINT error_type_logs_address_check,
    DROP COLUMN error_type,
    DROP COLUMN logs_address,
    DROP COLUMN owner,
    DROP COLUMN creation_date;

    DROP TABLE error_types;
$$) WHERE table_exists('public', 'error_types');
