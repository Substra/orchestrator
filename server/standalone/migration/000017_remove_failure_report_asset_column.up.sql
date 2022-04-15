SELECT execute($$
    CREATE TABLE error_types (
        error_type varchar(50) PRIMARY KEY
    );

    INSERT INTO error_types (error_type)
    VALUES ('ERROR_TYPE_BUILD'),
           ('ERROR_TYPE_EXECUTION'),
           ('ERROR_TYPE_INTERNAL');

    ALTER TABLE failure_reports
    ADD COLUMN error_type    varchar(50) REFERENCES error_types (error_type),
    ADD COLUMN logs_address  varchar(200) REFERENCES addressables (storage_address),
    ADD COLUMN owner         varchar(100),
    ADD COLUMN creation_date timestamptz,
    ADD CONSTRAINT failure_reports_owner_channel_fkey FOREIGN KEY (owner, channel) REFERENCES nodes (id, channel),
    ADD CONSTRAINT error_type_logs_address_check CHECK (
        (error_type = 'ERROR_TYPE_EXECUTION' AND logs_address IS NOT NULL) OR
        (error_type != 'ERROR_TYPE_EXECUTION' AND logs_address IS NULL)
    );

    INSERT INTO addressables(storage_address, checksum)
    SELECT asset -> 'logsAddress' ->> 'storageAddress',
           asset -> 'logsAddress' ->> 'checksum'
    FROM failure_reports
    WHERE (asset -> 'logsAddress') IS NOT NULL;

    UPDATE failure_reports
    SET error_type    = asset ->> 'errorType',
        logs_address  = asset -> 'logsAddress' ->> 'storageAddress',
        creation_date = (asset ->> 'creationDate')::timestamptz,
        owner         = asset ->> 'owner';

    ALTER TABLE failure_reports
    ALTER COLUMN error_type SET NOT NULL,
    ALTER COLUMN owner SET NOT NULL,
    ALTER COLUMN creation_date SET NOT NULL;

    ALTER TABLE failure_reports
    RENAME COLUMN compute_task_id TO compute_task_key;

    ALTER TABLE failure_reports
    RENAME CONSTRAINT failure_reports_compute_task_id_fkey TO failure_reports_compute_task_key_fkey;

    ALTER TABLE failure_reports
    DROP COLUMN asset;

    CREATE VIEW expanded_failure_reports AS
    SELECT compute_task_key,
           error_type,
           logs_address,
           a.checksum AS logs_checksum,
           creation_date,
           owner,
           channel
    FROM failure_reports
    LEFT JOIN addressables a ON failure_reports.logs_address = a.storage_address;
$$) WHERE NOT view_exists('public', 'expanded_failure_reports');
